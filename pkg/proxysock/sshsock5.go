// socks_ssh_forward.go
package proxysock

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"kotori/pkg/confopt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// readExact reads exactly n bytes or returns error
func readExact(conn net.Conn, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(conn, buf)
	return buf, err
}

func handleSocksConn(local net.Conn, sshClient *ssh.Client) {
	defer local.Close()
	logp := NewPrintLog("handleSocksConn", "")

	// SOCKS5 handshake
	hdr, err := readExact(local, 2)
	if err != nil {
		logp.Print("->handshake read error:", err)
		return
	}
	if hdr[0] != 0x05 {
		logp.Print("->unsupported socks version:", hdr[0])
		return
	}
	nmethods := int(hdr[1])
	if _, err = readExact(local, nmethods); err != nil {
		logp.Print("->read methods error:", err)
		return
	}
	// reply: no auth (0x00)
	if _, err = local.Write([]byte{0x05, 0x00}); err != nil {
		logp.Print("->write handshake reply error:", err)
		return
	}

	// read request
	reqHead, err := readExact(local, 4)
	if err != nil {
		logp.Print("->read request header error:", err)
		return
	}
	if reqHead[0] != 0x05 {
		logp.Print("->unsupported request version:", reqHead[0])
		return
	}
	cmd := reqHead[1]
	// rsv := reqHead[2]
	atyp := reqHead[3]

	if cmd != 0x01 {
		// only CONNECT supported
		// reply: command not supported (0x07)
		local.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		logp.Print("->unsupported socks command:", cmd)
		return
	}

	var host string
	switch atyp {
	case 0x01:
		// IPv4
		b, err := readExact(local, 4)
		if err != nil {
			logp.Print("->read ipv4 error:", err)
			return
		}
		host = net.IP(b).String()
	case 0x03:
		// domain
		lenb, err := readExact(local, 1)
		if err != nil {
			logp.Print("->read domain length error:", err)
			return
		}
		dlen := int(lenb[0])
		db, err := readExact(local, dlen)
		if err != nil {
			logp.Print("->read domain error:", err)
			return
		}
		host = string(db)
	case 0x04:
		// IPv6
		b, err := readExact(local, 16)
		if err != nil {
			logp.Print("->read ipv6 error:", err)
			return
		}
		host = net.IP(b).String()
	default:
		local.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		logp.Print("->unknown atyp:", atyp)
		return
	}
	// port
	pb, err := readExact(local, 2)
	if err != nil {
		logp.Print("->read port error:", err)
		return
	}
	port := binary.BigEndian.Uint16(pb)
	dest := fmt.Sprintf("%s:%d", host, port)
	logp.Print("->SOCKS CONNECT:", dest)

	// Use sshClient.Dial to create connection from remote side to dest
	remote, err := sshClient.Dial("tcp", dest)
	if err != nil {
		// reply: general failure
		local.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		logp.PrintF("->ssh dial to %s failed: %v", dest, err)
		return
	}
	// reply: success (bind addr 0.0.0.0:0)
	_, _ = local.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// pipe data both ways
	done := make(chan struct{}, 2)
	go func() {
		_, err = io.Copy(remote, local)
		remote.Close()
		done <- struct{}{}
		if err != nil {
			logp.Print("->Copy remote->local err:", err)
		}
	}()
	go func() {
		_, err = io.Copy(local, remote)
		local.Close()
		if err != nil {
			logp.Print("->Copy local->remote err:", err)
		}
		done <- struct{}{}
	}()
	<-done
	<-done
	logp.PrintF("->connection %s closed", dest)
}

// TIP: ssh -ND 22122
func RunSSHSock5(ctx context.Context, conf *confopt.Config, onlineChan chan string) error {
	proxyOpt := conf.SockProxy
	var (
		user        = proxyOpt.ServerUser
		server      = proxyOpt.ServerHost
		keyFile     = proxyOpt.ServerPriKey
		password    = proxyOpt.ServerPassword
		listenLocal = proxyOpt.Local
		insecure    = false
		keepAlive   = 6

		logp = NewPrintLog("RunSSHSock5", "")
	)

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if !insecure {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	sshConf := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         20 * time.Second,
	}
	if len(keyFile) > 0 {
		priKey, err := PublicKeyAuth(keyFile, proxyOpt.ServerPriPass)
		if err != nil {
			return errors.New("RunSSHSock5: load key failed:" + err.Error())
		}
		sshConf.Auth = []ssh.AuthMethod{
			priKey,
		}
	}

	logp.Print("-> dialing ssh start:", server)
	// 使用带 TCP 保活的 net.Dialer 替代默认拨号（解决底层 15 分钟 NAT/防火墙超时假死）
	dialer := &net.Dialer{
		Timeout:   20 * time.Second,
		KeepAlive: 15 * time.Second, // 操作系统内核每 15 秒发送一次 TCP 保活包
	}

	var err error
	tcpConn, err := dialer.DialContext(ctx, "tcp", server)
	if err != nil {
		onlineChan <- "RestartSSHSockProxy"
		logp.Print("connect tcp.Dial fail:", err)
		return errors.New("RunSSHSock5: tcp dial failed:" + err.Error())
	}

	// 将底层 TCP 连接包装为 SSH 连接
	conn, chans, reqs, err := ssh.NewClientConn(tcpConn, server, sshConf)
	if err != nil {
		tcpConn.Close()
		onlineChan <- "RestartSSHSockProxy"
		logp.Print("connect ssh.NewClientConn fail:", err)
		return errors.New("RunSSHSock5: ssh client conn failed:" + err.Error())
	}
	sshClient := ssh.NewClient(conn, chans, reqs)
	defer func() {
		sshClient.Close()
		if err != nil {
			onlineChan <- "RestartSSHSockProxy"
			logp.Print("connect ssh.Dial fail:", err)
		}

	}()
	logp.Print("ssh connection success to", server)

	ln, err := net.Listen("tcp", listenLocal)
	if err != nil {
		return errors.New("RunSSHSock5: listen: " + listenLocal + " | err: " + err.Error())
	}
	defer ln.Close()

	if keepAlive > 0 {
		go keepAliveSendReq(ctx, sshClient, ln, keepAlive)
	}

	logp.PrintF("SOCKS5 listening on %s (forward via %s) \n", listenLocal, server)
	go func() {
		time.Sleep(time.Second)
		onlineChan <- "RunProxyServer"
	}()
	go sock5Cancel(ctx, ln, sshClient)
	for {
		conn, err := ln.Accept()
		if err != nil {
			logp.Print("->accept connect error:", err)
			break
		}
		go handleSocksConn(conn, sshClient)
	}
	return err
}

func sock5Cancel(ctx context.Context, listen net.Listener, sshClient *ssh.Client) {
	for {
		select {
		case <-ctx.Done():
			listen.Close()
			sshClient.Close()
			log.Println("sock5Cancel: ctx.Done")
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func keepAliveSendReq(ctx context.Context, sshClient *ssh.Client, listen net.Listener, keepAlive int) {
	logp := NewPrintLog("keepAliveSendReq", "")
	ticker := time.NewTicker(time.Duration(keepAlive) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logp.Print("get ctx cancel(), keepAliveSendReq exit")
			return
		case <-ticker.C:
			// 发送标准的 OpenSSH 保活请求，取代错误的 tcpip-forward
			_, keepByte, err := sshClient.SendRequest("keepalive@openssh.com", true, nil)
			logp.Print("SendRequest:", string(keepByte), "| len:", len(keepByte))
			if err != nil {
				logp.Print("->Error: keepalive failed, closing connection:", err)
				// 保活失败时主动关闭连接与监听，释放 Accept 阻塞，让 RunSSHSock5 退出以触发重新连接
				sshClient.Close()
				listen.Close()
				return
			}
		}
	}
}

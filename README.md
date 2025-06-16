
# go-ssh-proxy-cli

### 代号 `Kotori`
> 功能不变, 引入`spf13/cobra`包, 让命令更现代化并且具有bash补全 <br/>
> v1版项目地址 `https://github.com/melty-blood/go-ssh-proxy`

> 1. 工具可以达到 `ssh -NL 20022:IP_ADDR:22 -J TypeMoon satsuki@SERVER_IP -p 8606` 这样的效果
> 2. 具有网络检测功能, 可以查看是否能到达某个ip和端口
> 3. 附加功能: 查看本地某张图片在其他目录是否也存在, 就是查重

------


```go

# linux
# -ldflags '-extldflags "-static"' 静态编译可以解决动态链接库依赖问题
CGO_ENABLED=0 GOARCH=amd64 go build -o ./kotori_proxy -a -ldflags '-extldflags "-static"' kotori.go

# windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./kotori_proxy.exe -a -ldflags '-extldflags "-static"' kotori.go

```

### bash自动补全
1. 将打包好的命令放入PATH环境变量中 `sudo cp kotori_proxy /usr/local/bin/`
2. 执行命令生成补全脚本 `kotori_proxy -c > /dev/shm/kotori_proxy_bash`
3. 将脚本文件放入补全目录, 并重新打开窗口尝试使用tab补全 `sudo mv /dev/shm/kotori_proxy_bash /etc/bash_completion.d/kotori_proxy_bash`

> Tip: 代码中根命令 `Use: "kotori_proxy"` 一定要和打包后的命令 `kotori_proxy` 一样. <br/>
> 执行命令默认检测`./conf/conf.yaml`配置文件是否存在. 后期考虑将默认配置文件放在`/etc`目录下.

### 小鸟(kotori) 镇楼
![./lovelive.jpg](./lovelive.jpg)


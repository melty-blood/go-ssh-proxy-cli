package svc

import (
	"fmt"
	"kotori/pkg/acgpic"
	"kotori/pkg/confopt"
	"kotori/pkg/helpers"
	"kotori/pkg/network"
	"kotori/pkg/proxysock"

	"github.com/spf13/cobra"
)

func CommandRoute(confPath string) {
	conf := confopt.ReadConf(confPath)
	defaultCmd := conf.DefaultCommand
	commandMap := map[string]func(conf *confopt.Config){
		"sshproxy": RunSSHProxyFunc,
		"nettouch": RunNetTouchFunc,
		"acgpic":   RunACGPicFunc,
	}

	funcVal, ok := commandMap[defaultCmd]
	if !ok {
		fmt.Println(helpers.GetFailPic(2))
		return
	}

	funcVal(conf)
}

func RunSSHProxyFunc(conf *confopt.Config) {
	proxysock.UseSSHFunc(conf)
}

func RunSSHProxy() *cobra.Command {
	var flagJson bool
	var flagWhat bool
	var flagConfig string

	var sshProxyCmd = &cobra.Command{
		Use:   "sshproxy [string to print!!]",
		Short: "ssh -NL PORT:SERVER_IP:PORT USER@IP -p PORT",
		Long:  "The tool can achieve something like 'ssh-NL 20022:IP_ADDR: 22-J TypeMoon satsuki@SERVER_IP -p 8606'.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// tempFlag := cmd.Flag("json")
			// fmt.Println("cmd.Flag-- json ", tempFlag.Value)
			conf := confopt.ReadConf(flagConfig)
			if flagWhat {
				fmt.Println("this flag is run `ssh -NL`, args from config.yaml")
				return
			}
			if flagJson {
				confopt.PrintConfJson(conf)
				return
			}
			proxysock.UseSSHFunc(conf)
		},
	}

	sshProxyCmd.Flags().BoolVarP(&flagJson, "json", "j", false, "print config with json")
	sshProxyCmd.Flags().BoolVarP(&flagWhat, "what", "w", false, "ssh proxy 'ssh -NL' command")
	sshProxyCmd.Flags().StringVarP(&flagConfig, "config", "f", "./conf/conf.yaml", "configure file, default file path ./conf/config.yaml")

	return sshProxyCmd
}

func RunNetTouchFunc(conf *confopt.Config) {
	fmt.Println("Tip: NetTouch Not supported yet!")
}

func RunNetTouch() *cobra.Command {
	var flagIp string
	var flagPort string
	var flagTimeOut int
	var flagVersion bool

	var netTouchCmd = &cobra.Command{
		Use:   "nettouch [must need flag]",
		Short: "network connect.",
		Long:  "Tool to test whether the network is connectable.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			netOpt := &network.NetTouchOpt{
				Ip:          flagIp,
				Port:        flagPort,
				Timeout:     flagTimeOut,
				ShowVersion: flagVersion,
			}
			network.NetCanTouch(netOpt)
		},
	}

	netTouchCmd.Flags().StringVarP(&flagIp, "ip", "i", "", "ip: 127.0.0.1")
	netTouchCmd.Flags().StringVarP(&flagPort, "port", "p", "", "port: 6666")
	netTouchCmd.Flags().IntVarP(&flagTimeOut, "timeout", "t", 6, "Timeout after a few seconds")
	netTouchCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Show version info")

	return netTouchCmd
}

func RunACGPicFunc(conf *confopt.Config) {
	targetImg := conf.AcgPic.TargetImg
	searchImgDir := conf.AcgPic.SearchImgDir
	threshold := conf.AcgPic.Threshold

	fmt.Println("----------------")
	fmt.Println("final parms: ", targetImg, searchImgDir, threshold)
	fmt.Println("----------------")
	acgpic.SearchPic(targetImg, searchImgDir, threshold)
}

func RunACGPic() *cobra.Command {
	var flagTargetImg string
	var flagSearchImgDir string
	var flagThreshold int
	var flagJson bool
	var flagConfig string

	var acgPicCmd = &cobra.Command{
		Use:   "acgpic [can use default flag]",
		Short: "Find the specified image.",
		Long:  "Search for similar images in the specified directory.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			conf := confopt.ReadConf(flagConfig)
			if flagJson {
				confopt.PrintConfJson(conf)
				return
			}
			targetImg := conf.AcgPic.TargetImg
			searchImgDir := conf.AcgPic.SearchImgDir
			threshold := conf.AcgPic.Threshold
			if len(flagTargetImg) > 0 {
				targetImg = flagTargetImg
			}
			if len(flagSearchImgDir) > 0 {
				searchImgDir = flagSearchImgDir
			}
			if flagThreshold > 0 {
				threshold = flagThreshold
			}

			fmt.Println("----------------")
			fmt.Println("final parms: ", targetImg, searchImgDir, threshold)
			fmt.Println("----------------")
			acgpic.SearchPic(targetImg, searchImgDir, threshold)
		},
	}

	acgPicCmd.Flags().StringVarP(&flagTargetImg, "target-img", "t", "", "need search image")
	acgPicCmd.Flags().StringVarP(&flagSearchImgDir, "search-img-dir", "s", "", "search image directory")
	acgPicCmd.Flags().IntVarP(&flagThreshold, "threshold", "T", 0, "threshold value, this is the similarity of the pictures")
	acgPicCmd.Flags().BoolVarP(&flagJson, "json", "j", false, "print config with json")
	acgPicCmd.Flags().StringVarP(&flagConfig, "config", "f", "./conf/conf.yaml", "configure file, default file path ./conf/config.yaml")

	return acgPicCmd
}

package main

import (
	"kotori/internal/svc"
	"kotori/pkg/confopt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var (
		flagConfig string
		completion bool
		err        error
	)
	sshProxyCmd := svc.RunSSHProxy()
	netTouchCmd := svc.RunNetTouch()
	acgPicCmd := svc.RunACGPic()
	grepCmd := svc.RunGrepPro()
	publishCmd := svc.RunPublishGit()

	var rootCmd = &cobra.Command{
		Use: "kotori_proxy",
		Run: func(cmd *cobra.Command, args []string) {
			// run default command
			if completion {
				cmd.GenBashCompletion(os.Stdout)
				return
			}
			// if nothing arg, run default command.
			svc.CommandRoute(flagConfig)
		},
	}

	// ./conf/conf.yaml
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "f", "", "configure file, default file path {APP_PATH}/conf/config.yaml")
	if flagConfig == "" {
		flagConfig, err = confopt.GetConfDir()
		if err != nil {
			log.Fatalln("get conf dir fail, err:", err)
		}
	}

	rootCmd.Flags().BoolVarP(&completion, "completion", "c", false, "Generate completion script")
	rootCmd.AddCommand(sshProxyCmd, netTouchCmd, acgPicCmd, grepCmd, publishCmd)
	rootCmd.Execute()
}

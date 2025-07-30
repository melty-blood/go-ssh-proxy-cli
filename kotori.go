package main

import (
	"kotori/internal/svc"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var flagConfig string
	var completion bool
	sshProxyCmd := svc.RunSSHProxy()
	netTouchCmd := svc.RunNetTouch()
	acgPicCmd := svc.RunACGPic()
	grepCmd := svc.RunGrepPro()

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
	rootCmd.Flags().StringVarP(&flagConfig, "config", "f", "./conf/conf.yaml", "configure file, default file path ./conf/config.yaml")
	rootCmd.Flags().BoolVarP(&completion, "completion", "c", false, "Generate completion script")
	rootCmd.AddCommand(sshProxyCmd, netTouchCmd, acgPicCmd, grepCmd)
	rootCmd.Execute()
}

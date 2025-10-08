/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	stopRpcServer        string
	stopRpcAdminPassword string
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if stopRpcServer == "" || stopRpcAdminPassword == "" {
			DebugWarn("ERROR", "--rpc-file= and --rpc-admin-password= cannot be empty")
			os.Exit(0)
		}

		DebugInfo("Stop ...", stopRpcServer)
		SetGrpcClient(stopRpcServer)
		gcAdminStop()
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.PersistentFlags().StringVar(&stopRpcServer, "rpc-server", "0.0.0.0:8282", "rpc file within address")
	stopCmd.PersistentFlags().StringVar(&stopRpcAdminPassword, "rpc-admin-password", "123", "rpc admin password for auth")
}

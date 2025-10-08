/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	IsDebug            bool
	IsAllowOverWrite   bool
	IsAllowUserKey     bool
	IsDisableDelete    bool
	IsDisableSet       bool
	MinFreeDiskSpaceMB uint64
	MaxUploadSizeMB    int64
	MaxUploadSize      int64
	DataDir            string
	AltDataDir         string
	AdminPassword      string
	AutoBackupDir      string
	AutoBackupEvery    string

	Host string
	Port string
)

var (
	pidFile string
	rpcFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zstdb",
	Short: "",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		DebugInfo("zstdb", "Thanks for choosing zstdb")
		BeforeStart()
	},
	Run: func(cmd *cobra.Command, args []string) {
		SaveCurrentPID()
		SaveCurrentAddr()
		BeforeGrpcStart()
		StartGrpcServer()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	rootCmd.PersistentFlags().BoolVar(&IsDebug, "debug", false, "if print debug info")
	rootCmd.PersistentFlags().BoolVar(&IsAllowOverWrite, "allow-overwrite", false, "if overwrite when data exists")
	rootCmd.PersistentFlags().BoolVar(&IsAllowUserKey, "allow-user-key", false, "if allow user-defined key")
	rootCmd.PersistentFlags().BoolVar(&IsDisableDelete, "disable-delete", false, "if disable user to delete data")
	rootCmd.PersistentFlags().BoolVar(&IsDisableSet, "disable-set", false, "if disable user to write data")
	rootCmd.PersistentFlags().Int64Var(&MaxUploadSizeMB, "max-upload-size-mb", 16, "Max Upload Size(16~1024MB), default: 16")
	rootCmd.PersistentFlags().StringVar(&AltDataDir, "alt-data-dir", "", "replace the env var zstdb_data")
	rootCmd.PersistentFlags().StringVar(&Host, "host", "0.0.0.0", "host, default: 0.0.0.0")
	rootCmd.PersistentFlags().StringVar(&Port, "port", "8282", "port, default: 8282")

	rootCmd.PersistentFlags().Uint64Var(&MinFreeDiskSpaceMB, "min-free-disk-space-mb", 4096,
		"disable-set=true if free space is less than this value, minimum: 4096")
	rootCmd.PersistentFlags().StringVar(&AdminPassword, "admin-password", "123", "password for rpc::admin")
	rootCmd.PersistentFlags().StringVar(&AutoBackupDir, "auto-backup-dir", "", "if set, run autobackup every hour")
	rootCmd.PersistentFlags().StringVar(&AutoBackupEvery, "auto-backup-every", "@every 1h",
		"scheduler, format: \"@every 15m\", \"@every 1h\", \"@every 1h30m\"")
}

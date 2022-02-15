package driverset

import (
	"github.com/spf13/cobra"
)

// RootCmd is the root command for driverset.
var RootCmd = &cobra.Command{
	Use:   "driverset",
	Short: "driverset provides tools for managing file-based driver settings",
}

var configFile string

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/usr/etc/driverset.yml",
		"Provide an alternative config file.")
	RootCmd.AddCommand(
		ReadCmd,
		SetCmd,
	)
}

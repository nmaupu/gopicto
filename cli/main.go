package cli

import (
	"github.com/spf13/cobra"
)

var (
	BuildDate          string
	ApplicationVersion string
)

const (
	AppName = "gopicto"

	ConfigFlag = "config"
	OutputFlag = "output"
)

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "Picto and associated word PDF generator",
}

func init() {
	rootCmd.PersistentFlags().StringP("verbosity", "v", "info", "Set log verbosity")
}

func Execute() error {
	return rootCmd.Execute()
}

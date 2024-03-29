package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of gopicto",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg(getVersionMessage())
	},
}

func getVersionMessage() string {
	return fmt.Sprintf("%s version %s (build date: %s)", AppName, ApplicationVersion, BuildDate)
}

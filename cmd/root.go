package cmd

import (
	"github.com/fieldju/awesom-o/cmd/potato"
	"github.com/fieldju/awesom-o/cmd/version"
	"github.com/fieldju/awesom-o/cmd/whoami"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verboseFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awesom-o",
	Short: "Greetings. I am the AWESOM-O 4000.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Add Base Commands
	rootCmd.AddCommand(version.VersionCmd)
	rootCmd.AddCommand(potato.PotatoFactCmd)
	rootCmd.AddCommand(whoami.WhoamiCmd)

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "show more details")
	rootCmd.PersistentPreRunE = configureLogging
}

func configureLogging(cmd *cobra.Command, args []string) error {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.TextFormatter{})
	return nil
}
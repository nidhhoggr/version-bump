package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joe-at-startupmedia/version-bump/v2/bump"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var acceptedArgs = []string{"major", "minor", "patch"}

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("bump [%s]", strings.Join(acceptedArgs, "|")),
	Short: "Bump a semantic version of the project",
	Long: `This application helps incrementing a semantic version of a project.
It can bump the version in multiple different files at once,
for example in package.json and a Dockerfile.`,
	ValidArgs: []string{"major", "minor", "patch"},
	Args:      cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			run(bump.StringToVersion(args[0]))
		} else {
			cmd.Help()
		}
	},
	Version: bump.Version,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

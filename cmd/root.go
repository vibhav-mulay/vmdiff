package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "vmdiff",
	Short:            "File diffing tool",
	PersistentPreRun: initLogging,
}

var (
	verbose bool
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose")
}

func initLogging(cmd *cobra.Command, args []string) {
	log.SetPrefix("vmdiff: ")
	if !verbose {
		log.SetOutput(io.Discard)
	}
}

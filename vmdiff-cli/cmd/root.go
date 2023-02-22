package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vibhav-mulay/vmdiff"

	"github.com/spf13/cobra"
)

// the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "vmdiff",
	Short:            "File diffing tool",
	PersistentPreRun: initLogging,
}

var (
	verbose int
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v",
		"Verbose mode, specific multiple times for increased verbosity")
}

func initLogging(cmd *cobra.Command, args []string) {
	log.SetPrefix("vmdiff-cli: ")
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	if verbose == 0 {
		log.SetOutput(io.Discard)
		vmdiff.DisableLogging()
	} else {
		level := vmdiff.INFO
		switch verbose {
		case 1:
			level = vmdiff.ERROR
		case 2:
			level = vmdiff.INFO
		case 3:
			level = vmdiff.DEBUG
		default:
			level = vmdiff.TRACE
		}

		vmdiff.SetDefaultLogLevel(level)
	}
}

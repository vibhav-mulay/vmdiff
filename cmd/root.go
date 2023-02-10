package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               "vmdiff",
	Short:             "File diffing tool",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


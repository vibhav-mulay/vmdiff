package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Generate delta file.",
	Long:  "Create a delta file describing changes in the new file.",
	RunE:  doDelta,
}

func init() {
	RootCmd.AddCommand(deltaCmd)
}

func doDelta(cmd *cobra.Command, args []string) error {
	fmt.Println("DELTA")

	return nil
}



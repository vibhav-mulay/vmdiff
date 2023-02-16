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

var deltaOpts = struct {
	infileStr    string
	sigFileStr   string
	deltaFileStr string
}{}

func init() {
	RootCmd.AddCommand(signatureCmd)
	deltaCmd.Flags().StringVarP(&deltaOpts.infileStr, "in", "i", "",
		"Input file path")
	deltaCmd.MarkFlagRequired("in")
	deltaCmd.Flags().StringVarP(&deltaOpts.sigFileStr, "signsture-file", "s", "",
		"Signature file path")
	deltaCmd.MarkFlagRequired("in")
	deltaCmd.Flags().StringVarP(&deltaOpts.deltaFileStr, "delta-file", "d", "",
		"Delta file path")
}

func init() {
	RootCmd.AddCommand(deltaCmd)
}

func doDelta(cmd *cobra.Command, args []string) error {
	fmt.Println("DELTA")

	return nil
}

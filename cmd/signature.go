package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "Generate signature file.",
	Long:  "Create a signature file with chunk and hash details of the input file.",
	RunE:  doSignature,
}

func init() {
	RootCmd.AddCommand(signatureCmd)
}

func doSignature(cmd *cobra.Command, args []string) error {
	fmt.Println("SIGNATURE")

	return nil
}



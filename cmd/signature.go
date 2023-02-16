package cmd

import (
	"io"
	"log"
	"os"

	"vmdiff/chunker"
	"vmdiff/internal"

	"github.com/spf13/cobra"
)

var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "Generate signature file.",
	Long:  "Create a signature file with chunk and hash details of the input file.",
	RunE:  doSignature,
}

var sigOpts = struct {
	infileStr  string
	sigFileStr string
	chunkerStr string
}{}

func init() {
	RootCmd.AddCommand(signatureCmd)
	signatureCmd.Flags().StringVarP(&sigOpts.infileStr, "in", "i", "",
		"Input file path")
	signatureCmd.MarkFlagRequired("in")
	signatureCmd.Flags().StringVarP(&sigOpts.sigFileStr, "signature-file", "s", "",
		"Signature file path")
	signatureCmd.Flags().StringVarP(&sigOpts.chunkerStr, "chunker", "c", "fastcdc",
		"Chunker to be used (fastcdc)")
}

func doSignature(cmd *cobra.Command, args []string) error {
	infile, outfile, err := doValidateFileParams()
	if err != nil {
		return err
	}

	log.Printf("Initializing chunker: %s\n", sigOpts.chunkerStr)
	chunker, err := chunker.GetChunker(sigOpts.chunkerStr, infile)
	if err != nil {
		return err
	}

	log.Printf("Generating signature\n")
	sfile, err := internal.GenerateSignature(chunker)
	if err != nil {
		return err
	}

	log.Printf("Writing signature to file\n")
	sfile.Dump(outfile)

	return nil
}

func doValidateFileParams() (io.Reader, io.Writer, error) {
	log.Printf("Opening input file: %s\n", sigOpts.infileStr)
	ifile, err := os.Open(sigOpts.infileStr)
	if err != nil {
		return nil, nil, err
	}

	ofile := os.Stdout
	if sigOpts.sigFileStr != "" {
		var err error

		log.Printf("Creating signature file: %s\n", sigOpts.sigFileStr)
		ofile, err = os.OpenFile(sigOpts.sigFileStr, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, err
		}
	}

	return ifile, ofile, nil
}

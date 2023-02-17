package cmd

import (
	"context"
	"log"
	"os"

	"vmdiff/chunker"
	"vmdiff/internal"

	"github.com/spf13/cobra"
)

type SignatureCmdOpenFiles struct {
	inFile  *os.File
	sigFile *os.File
}

var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "Generate signature file.",
	Long:  "Create a signature file with chunk and hash details of the input file.",
	RunE:  doSignature,
}

var sigOpts = struct {
	inFileStr  string
	sigFileStr string
	chunkerStr string
}{}

func init() {
	RootCmd.AddCommand(signatureCmd)
	signatureCmd.Flags().StringVarP(&sigOpts.inFileStr, "in", "i", "",
		"Input file path")
	signatureCmd.MarkFlagRequired("in")
	signatureCmd.Flags().StringVarP(&sigOpts.sigFileStr, "signature-file", "s", "",
		"Signature file path")
	signatureCmd.Flags().StringVarP(&sigOpts.chunkerStr, "chunker", "c", "fastcdc",
		"Chunker to be used (fastcdc, rabinfp)")
}

func doSignature(cmd *cobra.Command, args []string) error {
	files, err := validateSigCmdFileParams()
	if err != nil {
		return err
	}
	defer closeSigCmdOpenFiles(files)

	ctx := context.Background()

	log.Printf("Initializing chunker: %s\n", sigOpts.chunkerStr)
	chunker, err := chunker.GetChunker(sigOpts.chunkerStr, files.inFile)
	if err != nil {
		return err
	}

	log.Println("Generating signature")
	signature, err := internal.GenerateSignature(ctx, chunker)
	if err != nil {
		return err
	}

	log.Println("Writing signature to file")
	signature.Dump(ctx, files.sigFile)

	return nil
}

func validateSigCmdFileParams() (*SignatureCmdOpenFiles, error) {
	files := &SignatureCmdOpenFiles{}

	log.Printf("Opening input file: %s\n", sigOpts.inFileStr)
	ifile, err := os.Open(sigOpts.inFileStr)
	if err != nil {
		return files, err
	}
	files.inFile = ifile

	ofile := os.Stdout
	if sigOpts.sigFileStr != "" {
		var err error

		log.Printf("Creating signature file: %s\n", sigOpts.sigFileStr)
		ofile, err = os.OpenFile(sigOpts.sigFileStr, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return files, err
		}
	}
	files.sigFile = ofile

	return files, nil
}

func closeSigCmdOpenFiles(files *SignatureCmdOpenFiles) {
	if files.inFile != nil {
		files.inFile.Close()
	}

	if files.sigFile != nil {
		files.sigFile.Close()
	}
}

package cmd

import (
	"context"
	"log"
	"os"

	"vmdiff/internal"

	"github.com/spf13/cobra"
)

type DeltaCmdOpenFiles struct {
	inFile    *os.File
	sigFile   *os.File
	deltaFile *os.File
}

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Generate delta file.",
	Long:  "Create a delta file describing changes in the new file.",
	RunE:  doDelta,
}

var deltaOpts = struct {
	inFileStr    string
	sigFileStr   string
	deltaFileStr string
}{}

func init() {
	RootCmd.AddCommand(deltaCmd)
	deltaCmd.Flags().StringVarP(&deltaOpts.inFileStr, "in", "i", "",
		"Input file path")
	deltaCmd.MarkFlagRequired("in")
	deltaCmd.Flags().StringVarP(&deltaOpts.sigFileStr, "signature-file", "s", "",
		"Signature file path")
	deltaCmd.MarkFlagRequired("in")
	deltaCmd.Flags().StringVarP(&deltaOpts.deltaFileStr, "delta-file", "d", "",
		"Delta file path")
}

func doDelta(cmd *cobra.Command, args []string) error {
	files, err := validateDeltaCmdFileParams()
	if err != nil {
		return err
	}
	defer closeDeltaCmdOpenFiles(files)

	ctx := context.Background()

	log.Println("Loading signature")
	signature, err := internal.LoadSignature(ctx, files.sigFile)
	if err != nil {
		return err
	}

	log.Println("Generating delta")
	_, err = internal.GenerateDelta(ctx, files.inFile, signature, files.deltaFile)
	if err != nil {
		return err
	}

	return nil
}

func validateDeltaCmdFileParams() (*DeltaCmdOpenFiles, error) {
	files := &DeltaCmdOpenFiles{}

	log.Printf("Opening input file: %s\n", deltaOpts.inFileStr)
	ifile, err := os.Open(deltaOpts.inFileStr)
	if err != nil {
		return files, err
	}
	files.inFile = ifile

	log.Printf("Opening signature file: %s\n", deltaOpts.sigFileStr)
	sigfile, err := os.Open(deltaOpts.sigFileStr)
	if err != nil {
		return files, err
	}
	files.sigFile = sigfile

	ofile := os.Stdout
	if deltaOpts.deltaFileStr != "" {
		var err error

		log.Printf("Creating delta file: %s\n", deltaOpts.deltaFileStr)
		ofile, err = os.OpenFile(deltaOpts.deltaFileStr, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return files, err
		}
	}
	files.deltaFile = ofile

	return files, nil
}

func closeDeltaCmdOpenFiles(files *DeltaCmdOpenFiles) {
	if files.inFile != nil {
		files.inFile.Close()
	}

	if files.sigFile != nil {
		files.sigFile.Close()
	}

	if files.deltaFile != nil {
		files.deltaFile.Close()
	}
}

package cmd

import (
	"context"
	"log"
	"os"

	"vmdiff/internal"

	"github.com/spf13/cobra"
)

type PatchCmdOpenFiles struct {
	inFile    *os.File
	deltaFile *os.File
	outFile   *os.File
}

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch old file with delta to generate new file.",
	Long:  "Use the old file and delta file to regenerate the old file.",
	RunE:  doPatch,
}

var patchOpts = struct {
	inFileStr    string
	deltaFileStr string
	outFileStr   string
	dryRun       bool
}{}

func init() {
	RootCmd.AddCommand(patchCmd)
	patchCmd.Flags().StringVarP(&patchOpts.inFileStr, "in", "i", "",
		"Input file path")
	patchCmd.MarkFlagRequired("in")
	patchCmd.Flags().StringVarP(&patchOpts.deltaFileStr, "delta-file", "d", "",
		"Delta file path")
	patchCmd.MarkFlagRequired("delta-file")
	patchCmd.Flags().StringVarP(&patchOpts.outFileStr, "out", "o", "",
		"Output file path")
	patchCmd.Flags().BoolVarP(&patchOpts.dryRun, "dry-run", "x", false,
		"Dry Run. Changes in delta file are described")
}

func doPatch(cmd *cobra.Command, args []string) error {
	files, err := validatePatchCmdFileParams()
	if err != nil {
		return err
	}
	defer closePatchCmdOpenFiles(files)

	ctx := context.Background()

	patch := internal.NewDeltaPatcher(files.inFile, files.outFile, files.deltaFile, patchOpts.dryRun)

	log.Println("Patching delta")
	err = patch.PatchDelta(ctx)
	if err != nil {
		return err
	}

	return nil
}

func validatePatchCmdFileParams() (*PatchCmdOpenFiles, error) {
	files := &PatchCmdOpenFiles{}

	log.Printf("Opening input file: %s\n", patchOpts.inFileStr)
	ifile, err := os.Open(patchOpts.inFileStr)
	if err != nil {
		return files, err
	}
	files.inFile = ifile

	log.Printf("Opening delta file: %s\n", patchOpts.deltaFileStr)
	deltafile, err := os.Open(patchOpts.deltaFileStr)
	if err != nil {
		return files, err
	}
	files.deltaFile = deltafile

	ofile := os.Stdout
	if patchOpts.outFileStr != "" {
		var err error
		log.Printf("Creating output file: %s\n", patchOpts.outFileStr)
		ofile, err = os.OpenFile(patchOpts.outFileStr, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return files, err
		}
	}
	files.outFile = ofile

	return files, nil
}

func closePatchCmdOpenFiles(files *PatchCmdOpenFiles) {
	if files.inFile != nil {
		files.inFile.Close()
	}

	if files.deltaFile != nil {
		files.deltaFile.Close()
	}

	if files.outFile != nil {
		files.outFile.Close()
	}
}

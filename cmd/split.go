package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dyammarcano/qrfiletransfer/pkg/qrcode"
	"github.com/dyammarcano/qrfiletransfer/pkg/qrfiletransfer"
	"github.com/spf13/cobra"
)

var (
	splitInputFile string
	splitOutputDir string
	qrSize         int
	minQRSize      int
	maxQRSize      int
	autoAdjustSize bool
	recoveryLevel  string
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split a file into QR code images",
	Long: `Split a file into multiple QR code images stored in an output directory.

Example:
  qrfiletransfer split -i myfile.txt -o output_directory

This will split myfile.txt into multiple QR code images and store them in output_directory.
The QR codes can later be joined back into the original file using the join command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input file
		if splitInputFile == "" {
			fmt.Println("Error: input file is required")
			if err := cmd.Help(); err != nil {
				fmt.Printf("Error displaying help: %v\n", err)
			}
			os.Exit(1)
		}

		// Check if an input file exists
		if _, err := os.Stat(splitInputFile); os.IsNotExist(err) {
			fmt.Printf("Error: input file '%s' does not exist\n", splitInputFile)
			os.Exit(1)
		}

		// If the output directory is not specified, use a default
		if splitOutputDir == "" {
			// Use the input file name as the output directory name
			baseName := filepath.Base(splitInputFile)
			baseNameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
			splitOutputDir = fmt.Sprintf("%s_qrcodes", baseNameWithoutExt)
		}

		// Create an output directory if it doesn't exist
		if err := os.MkdirAll(splitOutputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Create QRFileTransfer instance
		qrft := qrfiletransfer.NewQRFileTransfer()

		// Set QR code options
		if qrSize > 0 {
			qrft.SetQRSize(qrSize)
		}

		if minQRSize > 0 {
			qrft.SetMinQRSize(minQRSize)
		}

		if maxQRSize > 0 {
			qrft.SetMaxQRSize(maxQRSize)
		}

		qrft.SetAutoAdjustQRSize(autoAdjustSize)

		// Set a recovery level
		var level qrcode.RecoveryLevel
		switch recoveryLevel {
		case "low":
			level = qrcode.Low
		case "medium":
			level = qrcode.Medium
		case "high":
			level = qrcode.High
		case "highest":
			level = qrcode.Highest
		default:
			level = qrcode.Medium
		}
		qrft.SetRecoveryLevel(level)

		// Split the file into QR codes
		fmt.Printf("Splitting file '%s' into QR codes in directory '%s'...\n", splitInputFile, splitOutputDir)
		if err := qrft.FileToQRCodes(splitInputFile, splitOutputDir); err != nil {
			fmt.Printf("Error splitting file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully split file into QR codes. QR codes are stored in '%s/qrcodes'\n", splitOutputDir)
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	// Add flags
	splitCmd.Flags().StringVarP(&splitInputFile, "input", "i", "", "Input file to split (required)")
	splitCmd.Flags().StringVarP(&splitOutputDir, "output", "o", "", "Output directory for QR codes (default: <filename>_qrcodes)")
	splitCmd.Flags().IntVarP(&qrSize, "size", "s", 0, "QR code size in pixels (default: 800)")
	splitCmd.Flags().IntVar(&minQRSize, "min-size", 0, "Minimum QR code size in pixels (default: 400)")
	splitCmd.Flags().IntVar(&maxQRSize, "max-size", 0, "Maximum QR code size in pixels (default: 1600)")
	splitCmd.Flags().BoolVar(&autoAdjustSize, "auto-adjust", true, "Automatically adjust QR code size based on data size")
	splitCmd.Flags().StringVarP(&recoveryLevel, "recovery", "r", "medium", "QR code recovery level (low, medium, high, highest)")
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"awesomeProjectQrFileTransfer/pkg/qrfiletransfer"
	"github.com/spf13/cobra"
)

var (
	joinInputDir   string
	joinOutputFile string
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join QR code images into a file",
	Long: `Join QR code images from an input directory back into the original file.

Example:
  awesomeProjectQrFileTransfer join -i input_directory -o output_file.txt

This will join the QR code images in input_directory back into the original file
and save it as output_file.txt.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input directory
		if joinInputDir == "" {
			fmt.Println("Error: input directory is required")
			cmd.Help()
			os.Exit(1)
		}

		// Check if input directory exists
		if _, err := os.Stat(joinInputDir); os.IsNotExist(err) {
			fmt.Printf("Error: input directory '%s' does not exist\n", joinInputDir)
			os.Exit(1)
		}

		// Check if the qrcodes directory exists inside the input directory
		qrcodesDir := filepath.Join(joinInputDir, "qrcodes")
		if _, err := os.Stat(qrcodesDir); os.IsNotExist(err) {
			fmt.Printf("Error: QR codes directory '%s' does not exist\n", qrcodesDir)
			os.Exit(1)
		}

		// Check if the data directory exists inside the input directory
		dataDir := filepath.Join(joinInputDir, "data")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			fmt.Printf("Error: data directory '%s' does not exist\n", dataDir)
			os.Exit(1)
		}

		// If output file is not specified, use a default
		if joinOutputFile == "" {
			// Use the input directory name as the output file name
			baseName := filepath.Base(joinInputDir)
			// Remove "_qrcodes" suffix if present
			if strings.HasSuffix(baseName, "_qrcodes") {
				baseName = baseName[:len(baseName)-len("_qrcodes")]
			}
			joinOutputFile = baseName + "_reconstructed"
		}

		// Create output directory if it doesn't exist
		outputDir := filepath.Dir(joinOutputFile)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Create QRFileTransfer instance
		qrft := qrfiletransfer.NewQRFileTransfer()

		// Join the QR codes into a file
		fmt.Printf("Joining QR codes from directory '%s' into file '%s'...\n", joinInputDir, joinOutputFile)
		if err := qrft.QRCodesToFile(joinInputDir, joinOutputFile); err != nil {
			fmt.Printf("Error joining QR codes: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully joined QR codes into file '%s'\n", joinOutputFile)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)

	// Add flags
	joinCmd.Flags().StringVarP(&joinInputDir, "input", "i", "", "Input directory containing QR codes (required)")
	joinCmd.Flags().StringVarP(&joinOutputFile, "output", "o", "", "Output file path (default: <dirname>_reconstructed)")
}

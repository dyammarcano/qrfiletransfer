package main

import (
	"awesomeProjectQrFileTransfer/pkg/qrcode"
	"awesomeProjectQrFileTransfer/pkg/qrfiletransfer"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Define command-line flags
	encodeCmd := flag.NewFlagSet("encode", flag.ExitOnError)
	encodeInput := encodeCmd.String("input", "", "Input file to encode (required)")
	encodeOutput := encodeCmd.String("output", "", "Output directory for QR codes (required)")
	encodeSize := encodeCmd.Int("size", 512, "Size of QR codes in pixels")
	encodeRecovery := encodeCmd.String("recovery", "medium", "QR code recovery level (low, medium, high, highest)")

	decodeCmd := flag.NewFlagSet("decode", flag.ExitOnError)
	decodeInput := decodeCmd.String("input", "", "Directory containing QR codes to decode (required)")
	decodeOutput := decodeCmd.String("output", "", "Output file path (required)")

	// Check if a subcommand is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  qrfiletransfer encode -input <file> -output <directory> [-size <pixels>] [-recovery <level>]")
		fmt.Println("  qrfiletransfer decode -input <directory> -output <file>")
		os.Exit(1)
	}

	// Parse the appropriate command
	switch os.Args[1] {
	case "encode":
		encodeCmd.Parse(os.Args[2:])
	case "decode":
		decodeCmd.Parse(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Usage:")
		fmt.Println("  qrfiletransfer encode -input <file> -output <directory> [-size <pixels>] [-recovery <level>]")
		fmt.Println("  qrfiletransfer decode -input <directory> -output <file>")
		os.Exit(1)
	}

	// Handle encode command
	if encodeCmd.Parsed() {
		// Validate required flags
		if *encodeInput == "" {
			fmt.Println("Error: -input flag is required")
			encodeCmd.PrintDefaults()
			os.Exit(1)
		}
		if *encodeOutput == "" {
			fmt.Println("Error: -output flag is required")
			encodeCmd.PrintDefaults()
			os.Exit(1)
		}

		// Check if input file exists
		if _, err := os.Stat(*encodeInput); os.IsNotExist(err) {
			fmt.Printf("Error: Input file %s does not exist\n", *encodeInput)
			os.Exit(1)
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(*encodeOutput, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Create QRFileTransfer instance
		qrft := qrfiletransfer.NewQRFileTransfer()

		// Set QR code size
		qrft.SetQRSize(*encodeSize)

		// Set recovery level
		var recoveryLevel qrcode.RecoveryLevel
		switch *encodeRecovery {
		case "low":
			recoveryLevel = qrcode.Low
		case "medium":
			recoveryLevel = qrcode.Medium
		case "high":
			recoveryLevel = qrcode.High
		case "highest":
			recoveryLevel = qrcode.Highest
		default:
			fmt.Printf("Unknown recovery level: %s. Using medium.\n", *encodeRecovery)
			recoveryLevel = qrcode.Medium
		}
		qrft.SetRecoveryLevel(recoveryLevel)

		// Convert file to QR codes
		fmt.Printf("Converting %s to QR codes in %s...\n", *encodeInput, *encodeOutput)
		if err := qrft.FileToQRCodes(*encodeInput, *encodeOutput); err != nil {
			fmt.Printf("Error converting file to QR codes: %v\n", err)
			os.Exit(1)
		}

		// Count the number of QR codes created
		qrDir := filepath.Join(*encodeOutput, "qrcodes")
		qrFiles, err := filepath.Glob(filepath.Join(qrDir, "*.png"))
		if err != nil {
			fmt.Printf("Error counting QR codes: %v\n", err)
		} else {
			fmt.Printf("Created %d QR codes in %s\n", len(qrFiles), qrDir)
		}

		fmt.Println("File successfully converted to QR codes!")
	}

	// Handle decode command
	if decodeCmd.Parsed() {
		// Validate required flags
		if *decodeInput == "" {
			fmt.Println("Error: -input flag is required")
			decodeCmd.PrintDefaults()
			os.Exit(1)
		}
		if *decodeOutput == "" {
			fmt.Println("Error: -output flag is required")
			decodeCmd.PrintDefaults()
			os.Exit(1)
		}

		// Check if input directory exists
		if _, err := os.Stat(*decodeInput); os.IsNotExist(err) {
			fmt.Printf("Error: Input directory %s does not exist\n", *decodeInput)
			os.Exit(1)
		}

		// Create output directory if it doesn't exist
		outputDir := filepath.Dir(*decodeOutput)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Create QRFileTransfer instance
		qrft := qrfiletransfer.NewQRFileTransfer()

		// Reconstruct file from QR codes
		fmt.Printf("Reconstructing file from QR codes in %s...\n", *decodeInput)
		if err := qrft.QRCodesToFile(*decodeInput, *decodeOutput); err != nil {
			fmt.Printf("Error reconstructing file from QR codes: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("File successfully reconstructed to %s!\n", *decodeOutput)
	}
}

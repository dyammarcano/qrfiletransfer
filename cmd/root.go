package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qrfiletransfer",
	Short: "A tool to transfer files using QR codes",
	Long: `QR File Transfer is a tool that allows you to transfer files using QR codes.

You can split a file into multiple QR code images and later join those QR codes
back into the original file. This is useful for transferring files between devices
that don't have a direct connection but can scan QR codes.

Use the 'split' command to split a file into QR codes, and the 'join' command
to join QR codes back into a file.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"awesomeProjectQrFileTransfer/pkg/qrcode"
	"awesomeProjectQrFileTransfer/pkg/qrfiletransfer"
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
	generateVideo  bool
	videoFPS       int
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split a file into QR code images",
	Long: `Split a file into multiple QR code images stored in an output directory.

Example:
  awesomeProjectQrFileTransfer split -i myfile.txt -o output_directory

This will split myfile.txt into multiple QR code images and store them in output_directory.
The QR codes can later be joined back into the original file using the join command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input file
		if splitInputFile == "" {
			fmt.Println("Error: input file is required")
			cmd.Help()
			os.Exit(1)
		}

		// Check if input file exists
		if _, err := os.Stat(splitInputFile); os.IsNotExist(err) {
			fmt.Printf("Error: input file '%s' does not exist\n", splitInputFile)
			os.Exit(1)
		}

		// If output directory is not specified, use a default
		if splitOutputDir == "" {
			// Use the input file name as the output directory name
			baseName := filepath.Base(splitInputFile)
			baseNameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
			splitOutputDir = fmt.Sprintf("%s_qrcodes", baseNameWithoutExt)
		}

		// Create output directory if it doesn't exist
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

		// Set recovery level
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

		// Generate video from QR codes if requested
		if generateVideo {
			fmt.Println("Generating video from QR codes...")

			// Check if ffmpeg is installed
			if err := checkFFmpegInstalled(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Generate video from QR codes
			qrDir := filepath.Join(splitOutputDir, "qrcodes")
			videoPath := filepath.Join(splitOutputDir, "qrcodes_video.mp4")

			if err := generateQRCodeVideo(qrDir, videoPath, videoFPS); err != nil {
				fmt.Printf("Error generating video: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Successfully generated video: %s\n", videoPath)
		}
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
	splitCmd.Flags().BoolVar(&generateVideo, "video", false, "Generate a video from QR codes using ffmpeg")
	splitCmd.Flags().IntVar(&videoFPS, "fps", 2, "Frames per second for the generated video (default: 2)")
}

// checkFFmpegInstalled checks if ffmpeg is installed on the system
func checkFFmpegInstalled() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg is not installed or not in PATH. Please install ffmpeg to use the video generation feature")
	}
	return nil
}

// generateQRCodeVideo generates a video from QR code images using ffmpeg
func generateQRCodeVideo(qrDir, videoPath string, fps int) error {
	// Get all PNG files in the QR codes directory
	files, err := filepath.Glob(filepath.Join(qrDir, "*.png"))
	if err != nil {
		return fmt.Errorf("failed to list QR code files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no QR code images found in %s", qrDir)
	}

	// Sort files to ensure they are processed in the correct order
	sort.Strings(files)

	// Create a temporary file with the list of images
	tempFile, err := os.CreateTemp("", "qrcodes_list_*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the list of files to the temporary file
	for _, file := range files {
		// Use file's absolute path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		// ffmpeg requires the file list to use the 'file' protocol
		_, err = tempFile.WriteString(fmt.Sprintf("file '%s'\n", absPath))
		if err != nil {
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
		// Set the duration for each image (in seconds)
		_, err = tempFile.WriteString(fmt.Sprintf("duration %f\n", 1.0/float64(fps)))
		if err != nil {
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Build the ffmpeg command
	cmd := exec.Command(
		"ffmpeg",
		"-y",           // Overwrite output file if it exists
		"-f", "concat", // Use concat demuxer
		"-safe", "0", // Don't require safe filenames
		"-i", tempFile.Name(), // Input file list
		"-vsync", "vfr", // Variable frame rate
		"-pix_fmt", "yuv420p", // Pixel format for compatibility
		"-c:v", "libx264", // Video codec
		videoPath, // Output file
	)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

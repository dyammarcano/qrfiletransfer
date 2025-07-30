package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var (
	generateInputDir string
	generateVideoFPS int
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a video from QR code images",
	Long: `Generate a video from QR code images using ffmpeg.

Example:
  awesomeProjectQrFileTransfer generate -i qrcodes_directory

This will generate a video from all QR code images in the specified directory.
The video will be saved in the same directory as "qrcodes_video.mp4".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input directory
		if generateInputDir == "" {
			fmt.Println("Error: input directory is required")
			if err := cmd.Help(); err != nil {
				fmt.Printf("Error displaying help: %v\n", err)
			}
			os.Exit(1)
		}

		// Check if the input directory exists
		if _, err := os.Stat(generateInputDir); os.IsNotExist(err) {
			fmt.Printf("Error: input directory '%s' does not exist\n", generateInputDir)
			os.Exit(1)
		}

		// Find the QR codes directory
		qrDir := generateInputDir
		// If the input is the parent directory, look for the qrcodes subdirectory
		qrcodesSubdir := filepath.Join(generateInputDir, "qrcodes")
		if _, err := os.Stat(qrcodesSubdir); err == nil {
			qrDir = qrcodesSubdir
		}

		fmt.Println("Generating video from QR codes...")

		// Check if ffmpeg is installed
		if err := checkFFmpegInstalled(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Generate video from QR codes
		videoPath := filepath.Join(filepath.Dir(qrDir), "qrcodes_video.mp4")
		if err := generateQRCodeVideo(qrDir, videoPath, generateVideoFPS); err != nil {
			fmt.Printf("Error generating video: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully generated video: %s\n", videoPath)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Add flags
	generateCmd.Flags().StringVarP(&generateInputDir, "input", "i", "", "Input directory containing QR codes (required)")
	generateCmd.Flags().IntVar(&generateVideoFPS, "fps", 5, "Frames per second for the generated video (default: 2)")
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
func generateQRCodeVideo(qrDir, videoPath string, fps int) (err error) {
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
	defer func() {
		removeErr := os.Remove(tempFile.Name())
		if removeErr != nil && err == nil {
			err = fmt.Errorf("failed to remove temporary file: %w", removeErr)
		}
	}()

	// Write the list of files to the temporary file
	for _, file := range files {
		// Use the file's absolute path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		// ffmpeg requires the file list to use the 'file' protocol
		_, err = fmt.Fprintf(tempFile, "file '%s'\n", absPath)
		if err != nil {
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
		// Set the duration for each image (in seconds)
		_, err = fmt.Fprintf(tempFile, "duration %f\n", 1.0/float64(fps))
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
		"-y",           // Overwrite an output file if it exists
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

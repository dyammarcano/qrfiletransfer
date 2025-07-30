package cmd

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"awesomeProjectQrFileTransfer/pkg/qrfiletransfer"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/spf13/cobra"
)

var (
	readInputVideo string
	readOutputFile string
	readTempDir    string
	readKeepFrames bool
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read QR codes from a video and reconstruct the file",
	Long: `Read QR codes from a video file and reconstruct the original file.

Example:
  awesomeProjectQrFileTransfer read -i qrcodes_video.mp4 -o reconstructed_file.txt

This will extract frames from the video, read QR codes from the frames,
and reconstruct the original file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input video
		if readInputVideo == "" {
			fmt.Println("Error: input video is required")
			if err := cmd.Help(); err != nil {
				fmt.Printf("Error displaying help: %v\n", err)
			}
			os.Exit(1)
		}

		// Check if the input video exists
		if _, err := os.Stat(readInputVideo); os.IsNotExist(err) {
			fmt.Printf("Error: input video '%s' does not exist\n", readInputVideo)
			os.Exit(1)
		}

		// If an output file is not specified, use a default
		if readOutputFile == "" {
			// Use the input video name as the output file name
			baseName := filepath.Base(readInputVideo)
			// Remove extension
			baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
			readOutputFile = fmt.Sprintf("%s_reconstructed", baseName)
		}

		// Create an output directory if it doesn't exist
		outputDir := filepath.Dir(readOutputFile)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// If a temp directory is not specified, create a temporary one
		if readTempDir == "" {
			var err error
			readTempDir, err = os.MkdirTemp("", "qrcode_frames_*")
			if err != nil {
				fmt.Printf("Error creating temporary directory: %v\n", err)
				os.Exit(1)
			}
			// Clean up the temporary directory if not keeping frames
			if !readKeepFrames {
				defer func() {
					if err := os.RemoveAll(readTempDir); err != nil {
						fmt.Printf("Warning: failed to remove temporary directory: %v\n", err)
					}
				}()
			}
		} else {
			// Create the specified temp directory if it doesn't exist
			if err := os.MkdirAll(readTempDir, 0755); err != nil {
				fmt.Printf("Error creating temporary directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Check if ffmpeg is installed
		if err := checkFFmpegInstalled(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Reading QR codes from video '%s'...\n", readInputVideo)

		// Extract frames from the video
		framesDir := filepath.Join(readTempDir, "frames")
		if err := os.MkdirAll(framesDir, 0755); err != nil {
			fmt.Printf("Error creating frames directory: %v\n", err)
			os.Exit(1)
		}

		if err := extractFramesFromVideo(readInputVideo, framesDir); err != nil {
			fmt.Printf("Error extracting frames: %v\n", err)
			os.Exit(1)
		}

		// Create directories for QR code data
		qrcodesDir := filepath.Join(readTempDir, "qrcodes")
		if err := os.MkdirAll(qrcodesDir, 0755); err != nil {
			fmt.Printf("Error creating QR codes directory: %v\n", err)
			os.Exit(1)
		}

		dataDir := filepath.Join(readTempDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			fmt.Printf("Error creating data directory: %v\n", err)
			os.Exit(1)
		}

		// Read QR codes from frames
		if err := readQRCodesFromFrames(framesDir, dataDir); err != nil {
			fmt.Printf("Error reading QR codes: %v\n", err)
			os.Exit(1)
		}

		// Copy the extracted frames to the qrcodes directory for reference
		// This is not strictly necessary for reconstruction but helps with debugging
		if readKeepFrames {
			fmt.Println("Copying extracted frames to qrcodes directory for reference...")
			frames, err := filepath.Glob(filepath.Join(framesDir, "*.png"))
			if err != nil {
				fmt.Printf("Warning: failed to list frames for copying: %v\n", err)
			} else if len(frames) > 0 {
				copiedFrames := 0
				for i, frame := range frames {
					// Only copy a reasonable number of frames to avoid excessive disk usage
					if i >= 10 {
						break
					}
					destPath := filepath.Join(qrcodesDir, filepath.Base(frame))
					if err := copyFile(frame, destPath); err != nil {
						fmt.Printf("Warning: failed to copy frame %s: %v\n", frame, err)
					} else {
						copiedFrames++
					}
				}
				fmt.Printf("Copied %d frames to %s for reference\n", copiedFrames, qrcodesDir)
			}
		}

		// Create QRFileTransfer instance
		qrft := qrfiletransfer.NewQRFileTransfer()

		// Reconstruct the file from QR codes
		fmt.Printf("Reconstructing file from QR codes...\n")
		if err := qrft.QRCodesToFile(readTempDir, readOutputFile); err != nil {
			fmt.Printf("Error reconstructing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully reconstructed file: %s\n", readOutputFile)
		if readKeepFrames {
			fmt.Printf("Extracted frames and intermediate files are kept in: %s\n", readTempDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	// Add flags
	readCmd.Flags().StringVarP(&readInputVideo, "input", "i", "", "Input video file containing QR codes (required)")
	readCmd.Flags().StringVarP(&readOutputFile, "output", "o", "", "Output file path (default: <videoname>_reconstructed)")
	readCmd.Flags().StringVarP(&readTempDir, "temp", "t", "", "Temporary directory for extracted frames (default: system temp)")
	readCmd.Flags().BoolVarP(&readKeepFrames, "keep", "k", false, "Keep extracted frames and intermediate files")
}

// extractFramesFromVideo extracts frames from a video using ffmpeg
func extractFramesFromVideo(videoPath, outputDir string) error {
	// Build the ffmpeg command to extract frames
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vsync", "0",
		"-q:v", "2", // High quality
		filepath.Join(outputDir, "frame_%04d.png"),
	)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// readQRCodesFromFrames reads QR codes from image frames and saves the data
func readQRCodesFromFrames(framesDir, dataDir string) error {
	// Get all PNG files in the frames directory
	frames, err := filepath.Glob(filepath.Join(framesDir, "*.png"))
	if err != nil {
		return fmt.Errorf("failed to list frame files: %w", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no frames found in %s", framesDir)
	}

	fmt.Printf("Processing %d frames...\n", len(frames))

	// Track successfully processed frames and unique chunks
	processedFrames := 0
	processedChunks := make(map[string]bool)

	// Process each frame
	for i, framePath := range frames {
		// Read QR code from the frame
		data, err := readQRCodeFromImage(framePath)
		if err != nil {
			// Just log the error and continue with the next frame
			fmt.Printf("Warning: failed to read QR code from frame %s: %v\n", framePath, err)
			continue
		}

		// Generate a simple hash of the data to detect duplicates
		// This is a simple approach - in a production system, you might want to use a more robust method
		dataHash := fmt.Sprintf("%x", data[:minV(len(data), 20)])

		// Skip if we've already processed this chunk (duplicate frame)
		if processedChunks[dataHash] {
			fmt.Printf("Info: skipping duplicate QR code in frame %s\n", framePath)
			continue
		}

		// Mark this chunk as processed
		processedChunks[dataHash] = true

		// Save the data to a file
		dataFilePath := filepath.Join(dataDir, fmt.Sprintf("chunk_%04d.dat", processedFrames))
		if err := os.WriteFile(dataFilePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write data to file %s: %w", dataFilePath, err)
		}

		processedFrames++
		fmt.Printf("Processed frame %d/%d (found %d unique QR codes)\r", i+1, len(frames), processedFrames)
	}
	fmt.Println() // Print a newline after the progress indicator

	if processedFrames == 0 {
		return fmt.Errorf("no valid QR codes found in any frames")
	}

	fmt.Printf("Successfully extracted %d unique QR codes from %d frames\n", processedFrames, len(frames))
	return nil
}

// minV returns the minimum of two integers
func minV(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to the destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}

// readQRCodeFromImage reads a QR code from an image file
func readQRCodeFromImage(imagePath string) ([]byte, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close image file: %v\n", err)
		}
	}(file)

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create a binary bitmap from the image
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary bitmap: %w", err)
	}

	// Create a QR code reader
	reader := qrcode.NewQRCodeReader()

	// Try to decode the QR code
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QR code: %w", err)
	}

	// Get the text from the result
	text := result.GetText()

	// The QR code content is expected to be base64 encoded
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return data, nil
}

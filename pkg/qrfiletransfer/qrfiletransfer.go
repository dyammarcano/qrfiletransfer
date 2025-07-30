package qrfiletransfer

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"awesomeProjectQrFileTransfer/pkg/qrcode"
	"awesomeProjectQrFileTransfer/pkg/split"
)

// QRFileTransfer handles the conversion of files to QR codes and back
type QRFileTransfer struct {
	splitter *split.Split
	// Maximum chunk size in bytes (considering QR code capacity)
	// Version 40 with Low recovery level can encode up to 2953 bytes
	// Using a slightly smaller value to be safe
	maxChunkSize int
	// QR code recovery level
	recoveryLevel qrcode.RecoveryLevel
	// QR code size in pixels
	qrSize int
	// Minimum QR code size in pixels
	minQRSize int
	// Maximum QR code size in pixels
	maxQRSize int
	// Enable automatic QR size adjustment based on content
	autoAdjustQRSize bool
}

// NewQRFileTransfer creates a new QRFileTransfer instance
func NewQRFileTransfer() *QRFileTransfer {
	return &QRFileTransfer{
		splitter:         split.NewSplit(),
		maxChunkSize:     2000, // Using a conservative value to ensure QR codes can be generated
		recoveryLevel:    qrcode.Medium,
		qrSize:           800,  // Default QR code size in pixels
		minQRSize:        400,  // Minimum QR code size in pixels
		maxQRSize:        1600, // Maximum QR code size in pixels
		autoAdjustQRSize: true, // Enable automatic QR size adjustment by default
	}
}

// SetRecoveryLevel sets the QR code recovery level
func (q *QRFileTransfer) SetRecoveryLevel(level qrcode.RecoveryLevel) {
	q.recoveryLevel = level
}

// SetQRSize sets the QR code size in pixels
func (q *QRFileTransfer) SetQRSize(size int) {
	q.qrSize = size
}

// SetMinQRSize sets the minimum QR code size in pixels
func (q *QRFileTransfer) SetMinQRSize(size int) {
	q.minQRSize = size
}

// SetMaxQRSize sets the maximum QR code size in pixels
func (q *QRFileTransfer) SetMaxQRSize(size int) {
	q.maxQRSize = size
}

// SetAutoAdjustQRSize enables or disables automatic QR size adjustment
func (q *QRFileTransfer) SetAutoAdjustQRSize(enable bool) {
	q.autoAdjustQRSize = enable
}

// calculateOptimalQRSize calculates the optimal QR code size in pixels based on the chunk size
// It estimates the QR code version based on the chunk size and then calculates an appropriate pixel size
func (q *QRFileTransfer) calculateOptimalQRSize(chunkSize int) int {
	// Base64 encoding increases the size by approximately 4/3
	encodedSize := int(float64(chunkSize) * 1.34)

	// Add some overhead for the chunk identifier and format
	// "Chunk: name\nData: base64data"
	encodedSize += 20

	// Estimate QR code version based on data size and recovery level
	// These are rough estimates based on QR code capacity
	var estimatedVersion int

	// Capacity in bytes for different versions with Medium recovery level
	// Version 1: ~16 bytes, Version 10: ~271 bytes, Version 20: ~858 bytes, Version 30: ~1732 bytes, Version 40: ~2953 bytes
	switch {
	case encodedSize < 16:
		estimatedVersion = 1
	case encodedSize < 100:
		estimatedVersion = 5
	case encodedSize < 271:
		estimatedVersion = 10
	case encodedSize < 500:
		estimatedVersion = 15
	case encodedSize < 858:
		estimatedVersion = 20
	case encodedSize < 1300:
		estimatedVersion = 25
	case encodedSize < 1732:
		estimatedVersion = 30
	case encodedSize < 2300:
		estimatedVersion = 35
	default:
		estimatedVersion = 40
	}

	// Calculate module size (number of modules in the QR code)
	// Formula: 21 + (version-1)*4
	moduleSize := 21 + (estimatedVersion-1)*4

	// Calculate pixel size based on module size
	// We want each module to be at least 2 pixels for readability
	// But we also want to keep the QR code size reasonable
	pixelSize := moduleSize * 4

	// Ensure the size is within the min and max bounds
	if pixelSize < q.minQRSize {
		pixelSize = q.minQRSize
	} else if pixelSize > q.maxQRSize {
		pixelSize = q.maxQRSize
	}

	return pixelSize
}

// FileToQRCodes converts a file to a series of QR codes
// Parameters:
//   - filePath: Path to the file to convert
//   - outDir: Directory to store the QR codes
//
// Returns an error if any part of the process fails.
func (q *QRFileTransfer) FileToQRCodes(filePath string, outDir string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a temporary directory for chunks
	tempDir := filepath.Join(outDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Calculate number of chunks based on file size and max chunk size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()

	// Calculate number of chunks based on file size
	// For larger files, we need more chunks to ensure each chunk is small enough for QR encoding
	var numChunks int
	switch {
	case fileSize <= 1000:
		// For small files, use at least 2 chunks
		numChunks = 2
	case fileSize <= 5000:
		// For medium files, ensure chunks are around 1000 bytes or less
		numChunks = int(fileSize/1000) + 1
	case fileSize <= 20000:
		// For larger files, ensure chunks are around 800 bytes or less
		numChunks = int(fileSize/800) + 1
	default:
		// For very large files, ensure chunks are around 500 bytes or less
		numChunks = int(fileSize/500) + 1
	}

	// Update maxChunkSize based on the calculated number of chunks
	// This ensures that the chunks are properly sized for QR encoding
	q.maxChunkSize = int(fileSize/int64(numChunks)) + 1

	// Split the file into chunks
	if err := q.splitter.SplitFile(file, tempDir, numChunks); err != nil {
		return fmt.Errorf("failed to split file: %w", err)
	}

	// Create output directory for QR codes
	qrDir := filepath.Join(outDir, "qrcodes")
	if err := os.MkdirAll(qrDir, 0755); err != nil {
		return fmt.Errorf("failed to create QR codes directory: %w", err)
	}

	// Create output directory for raw data
	dataDir := filepath.Join(outDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Get all chunk files
	chunkFiles, err := filepath.Glob(filepath.Join(tempDir, "*.part"))
	if err != nil {
		return fmt.Errorf("failed to list chunk files: %w", err)
	}

	// Also include the first chunk which has a .tmp extension
	firstChunk, err := filepath.Glob(filepath.Join(tempDir, "*.tmp"))
	if err != nil {
		return fmt.Errorf("failed to find first chunk: %w", err)
	}

	if len(firstChunk) > 0 {
		chunkFiles = append(firstChunk, chunkFiles...)
	}

	// Convert each chunk to a QR code and store raw data
	for _, chunkPath := range chunkFiles {
		// Read the chunk
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk %s: %w", chunkPath, err)
		}

		// Get the base name of the chunk file
		baseName := filepath.Base(chunkPath)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// Create QR code file name with the same naming convention
		qrFileName := baseNameWithoutExt + ".png"
		qrFilePath := filepath.Join(qrDir, qrFileName)

		// Create data file name with the same naming convention
		dataFileName := baseNameWithoutExt + ".dat"
		dataFilePath := filepath.Join(dataDir, dataFileName)

		// Create a QR code from the chunk data
		// For binary data, we need to use a string representation
		// This is a limitation of the QR code package
		// Encode the binary data as base64 string
		encodedData := base64.StdEncoding.EncodeToString(chunkData)
		qrContent := fmt.Sprintf("Chunk: %s\nData: %s", baseNameWithoutExt, encodedData)
		qrCode, err := qrcode.New(qrContent, q.recoveryLevel)
		if err != nil {
			return fmt.Errorf("failed to create QR code for chunk %s: %w", chunkPath, err)
		}

		// Determine the QR code size to use
		qrSize := q.qrSize
		if q.autoAdjustQRSize {
			// Calculate optimal QR code size based on chunk size
			qrSize = q.calculateOptimalQRSize(len(chunkData))
		}

		// Save the QR code to a file
		if err := qrCode.WriteFile(qrSize, qrFilePath); err != nil {
			return fmt.Errorf("failed to write QR code to file %s: %w", qrFilePath, err)
		}

		// Save the raw data to a file
		if err := os.WriteFile(dataFilePath, chunkData, 0644); err != nil {
			return fmt.Errorf("failed to write data to file %s: %w", dataFilePath, err)
		}
	}

	// Clean up temporary directory
	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("failed to clean up temporary directory: %w", err)
	}

	return nil
}

// QRCodesToFile reconstructs a file from a series of QR codes and their associated data files
// Parameters:
//   - inDir: Directory containing the QR codes and data files
//   - outFilePath: Path to save the reconstructed file
//
// Returns an error if any part of the process fails.
func (q *QRFileTransfer) QRCodesToFile(inDir string, outFilePath string) error {
	// Create a temporary directory for chunks
	tempDir := filepath.Join(inDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Get all data files
	dataDir := filepath.Join(inDir, "data")
	dataFiles, err := filepath.Glob(filepath.Join(dataDir, "*.dat"))
	if err != nil {
		return fmt.Errorf("failed to list data files: %w", err)
	}

	if len(dataFiles) == 0 {
		return fmt.Errorf("no data files found in %s", dataDir)
	}

	// Process each data file
	for _, dataFilePath := range dataFiles {
		// Read the data file
		chunkData, err := os.ReadFile(dataFilePath)
		if err != nil {
			return fmt.Errorf("failed to read data file %s: %w", dataFilePath, err)
		}

		// Get the base name of the data file
		baseName := filepath.Base(dataFilePath)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// All chunks should have .part extension
		// The first chunk is identified by its index (0), not by its extension
		chunkFilePath := filepath.Join(tempDir, baseNameWithoutExt+".part")

		// Write the chunk data to a file
		if err := os.WriteFile(chunkFilePath, chunkData, 0644); err != nil {
			return fmt.Errorf("failed to write chunk to file %s: %w", chunkFilePath, err)
		}
	}

	// Merge the chunks to reconstruct the original file
	if err := q.splitter.MergeFile(tempDir); err != nil {
		return fmt.Errorf("failed to merge chunks: %w", err)
	}

	// Find the reconstructed file in the temp directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temporary directory: %w", err)
	}

	var reconstructedFile string
	for _, file := range files {
		if !file.IsDir() && !strings.HasSuffix(file.Name(), ".part") && !strings.HasSuffix(file.Name(), ".tmp") {
			reconstructedFile = filepath.Join(tempDir, file.Name())
			break
		}
	}

	if reconstructedFile == "" {
		return fmt.Errorf("reconstructed file not found")
	}

	// Copy the reconstructed file to the output path
	srcFile, err := os.Open(reconstructedFile)
	if err != nil {
		return fmt.Errorf("failed to open reconstructed file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(outFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy reconstructed file: %w", err)
	}

	return nil
}

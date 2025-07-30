# QR File Transfer

QR File Transfer is a tool that allows you to transfer files using QR codes. You can split a file into multiple QR code images and later join those QR codes back into the original file. This is useful for transferring files between devices that don't have a direct connection but can scan QR codes.

## Features

- **Split files into QR codes**: Convert any file into a series of QR code images
- **Join QR codes back into files**: Reconstruct the original file from QR code images
- **Generate videos from QR codes**: Create videos from QR code images for easier transfer
- **Customizable QR codes**: Adjust QR code size, recovery level, and other parameters
- **Automatic size adjustment**: Optimize QR code size based on data content

## Installation

### Prerequisites

- Go 1.16 or higher
- For video generation: ffmpeg must be installed and available in your PATH

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/dyammarcano/qrfiletransfer.git
   cd qrfiletransfer
   ```

2. Build the application:
   ```
   go build -o qrfiletransfer
   ```

3. (Optional) Install the application:
   ```
   go install
   ```

## Usage

### Split a file into QR codes

```
qrfiletransfer split -i <input_file> -o <output_directory>
```

This will split the input file into multiple QR code images and store them in the specified output directory. If no output directory is specified, a directory named `<filename>_qrcodes` will be created.

#### Options

- `-i, --input`: Input file to split (required)
- `-o, --output`: Output directory for QR codes (default: `<filename>_qrcodes`)
- `-s, --size`: QR code size in pixels (default: 800)
- `--min-size`: Minimum QR code size in pixels (default: 400)
- `--max-size`: Maximum QR code size in pixels (default: 1600)
- `--auto-adjust`: Automatically adjust QR code size based on data size (default: true)
- `-r, --recovery`: QR code recovery level (low, medium, high, highest) (default: medium)

### Join QR codes into a file

```
qrfiletransfer join -i <input_directory> -o <output_file>
```

This will join the QR code images in the input directory back into the original file and save it as the specified output file. If no output file is specified, a file named `<dirname>_reconstructed` will be created.

#### Options

- `-i, --input`: Input directory containing QR codes (required)
- `-o, --output`: Output file path (default: `<dirname>_reconstructed`)

### Generate a video from QR codes

```
qrfiletransfer generate -i <input_directory> --fps <frames_per_second>
```

This will generate a video from all QR code images in the specified directory. The video will be saved in the same directory as "qrcodes_video.mp4". This feature requires ffmpeg to be installed.

#### Options

- `-i, --input`: Input directory containing QR codes (required)
- `--fps`: Frames per second for the generated video (default: 2)

## Examples

### Basic workflow

1. Split a file into QR codes:
   ```
   qrfiletransfer split -i document.pdf -o document_qrcodes
   ```

2. Transfer the QR codes to another device (by scanning them or transferring the images)

3. Join the QR codes back into the original file:
   ```
   qrfiletransfer join -i document_qrcodes -o document_restored.pdf
   ```

### Creating a video for easier transfer

1. Split a file into QR codes:
   ```
   qrfiletransfer split -i document.pdf
   ```

2. Generate a video from the QR codes:
   ```
   qrfiletransfer generate -i document.pdf_qrcodes --fps 1
   ```

3. Play the video on one device and scan the QR codes with another device

4. Join the scanned QR codes back into the original file:
   ```
   qrfiletransfer join -i scanned_qrcodes -o document_restored.pdf
   ```

## How It Works

QR File Transfer works by:

1. Splitting the input file into manageable chunks
2. Encoding each chunk as a QR code
3. Storing metadata about the file in additional QR codes
4. When joining, it decodes the QR codes and reassembles the original file

The tool uses error correction in QR codes to ensure reliable data transfer even if the QR code is partially damaged or difficult to scan.

## License

This project's license is not specified in the current state.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
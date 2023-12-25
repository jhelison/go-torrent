package filesystem

import (
	"os"
)

// CreateFileWithSize creates a new file with size filled empty 0 bytes
// Since we will be writing chunks, we don't close the file
func CreateFileWithSize(filename string, size int64) (*os.File, error) {
	// Create the file or truncate the existing one
	file, err := os.Create(filename)
	if err != nil {
		return file, err
	}

	// Crete the empty content for the file
	data := make([]byte, size)

	// Write to the file
	_, err = file.Write(data)
	if err != nil {
		return file, err
	}
	return file, nil
}

// WriteFileChunk write to a file a buffer with a offset
func WriteFileChunk(file *os.File, buffer []byte, offset int64) error {
	// Go to the offset
	_, err := file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Write to the file
	_, err = file.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}

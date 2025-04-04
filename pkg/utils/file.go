package utils

import "os"

// CheckIfFileExists checks if a file exists at the given path.
// It returns true if the file exists and is not a directory, false otherwise.
func CheckIfFileExists(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// Other errors occurred (e.g., permission issues)
		return false
	}
	// Check if it's a regular file (not a directory)
	return !fileInfo.IsDir()
}

package providers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// NFSConfig holds configuration for NFS provider
type NFSConfig struct {
	BasePath   string // NFS mount path, e.g., "/app/shared-files"
	BaseURL    string // Base URL for serving files, e.g., "http://your-domain.com/files"
	ServerIP   string // NFS server IP for direct access
	ServerPath string // Server-side path for direct operations
}

// NFSProvider implements file storage using NFS
type NFSProvider struct {
	config *NFSConfig
}

// NewNFSProvider creates a new NFS provider
func NewNFSProvider(config *NFSConfig) (*NFSProvider, error) {
	// Validate configuration
	if config.BasePath == "" {
		return nil, fmt.Errorf("basePath is required")
	}
	
	// Ensure base directory exists
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"uploads", "processed", "temp"}
	for _, subdir := range subdirs {
		dir := filepath.Join(config.BasePath, subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &NFSProvider{
		config: config,
	}, nil
}

// UploadFile uploads a file to NFS storage
func (n *NFSProvider) UploadFile(fileBuffer []byte, fileName string) error {
	// Determine subdirectory based on file type/purpose
	subdir := "uploads"
	if filepath.Ext(fileName) != "" {
		// You can add logic here to categorize files
		// For now, all uploads go to uploads directory
	}

	filePath := filepath.Join(n.config.BasePath, subdir, fileName)
	
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := ioutil.WriteFile(filePath, fileBuffer, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetSignedUrl returns a URL for accessing the file
func (n *NFSProvider) GetSignedUrl(fileName string) (string, error) {
	// For NFS, we return a direct HTTP URL served by your web server
	// You'll need to implement file serving in your Go API
	
	if n.config.BaseURL == "" {
		return "", fmt.Errorf("baseURL not configured for URL generation")
	}

	// Check if file exists
	if !n.FileExists(fileName) {
		return "", fmt.Errorf("file not found: %s", fileName)
	}

	// Return direct URL - your web server will handle serving these files
	url := fmt.Sprintf("%s/api/files/%s", n.config.BaseURL, fileName)
	return url, nil
}

// DeleteFile deletes a file from NFS storage
func (n *NFSProvider) DeleteFile(fileName string) error {
	// Try to find file in any subdirectory
	subdirs := []string{"uploads", "processed", "temp"}
	
	for _, subdir := range subdirs {
		filePath := filepath.Join(n.config.BasePath, subdir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			return os.Remove(filePath)
		}
	}

	return fmt.Errorf("file not found: %s", fileName)
}

// FileExists checks if a file exists in NFS storage
func (n *NFSProvider) FileExists(fileName string) bool {
	subdirs := []string{"uploads", "processed", "temp"}
	
	for _, subdir := range subdirs {
		filePath := filepath.Join(n.config.BasePath, subdir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			return true
		}
	}
	
	return false
}

// GetFilePath returns the full path to a file
func (n *NFSProvider) GetFilePath(fileName string, subdir string) string {
	if subdir == "" {
		subdir = "uploads"
	}
	return filepath.Join(n.config.BasePath, subdir, fileName)
}

// MoveFile moves a file from one location to another within NFS
func (n *NFSProvider) MoveFile(srcFileName, destFileName, destSubdir string) error {
	// Find source file
	var srcPath string
	subdirs := []string{"uploads", "processed", "temp"}
	
	for _, subdir := range subdirs {
		testPath := filepath.Join(n.config.BasePath, subdir, srcFileName)
		if _, err := os.Stat(testPath); err == nil {
			srcPath = testPath
			break
		}
	}
	
	if srcPath == "" {
		return fmt.Errorf("source file not found: %s", srcFileName)
	}

	// Destination path
	if destSubdir == "" {
		destSubdir = "processed"
	}
	destPath := filepath.Join(n.config.BasePath, destSubdir, destFileName)
	
	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move file
	return os.Rename(srcPath, destPath)
}

// ListFiles lists files in a specific subdirectory
func (n *NFSProvider) ListFiles(subdir string) ([]string, error) {
	if subdir == "" {
		subdir = "uploads"
	}

	dirPath := filepath.Join(n.config.BasePath, subdir)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// GetFileInfo returns file information
func (n *NFSProvider) GetFileInfo(fileName string) (os.FileInfo, error) {
	subdirs := []string{"uploads", "processed", "temp"}
	
	for _, subdir := range subdirs {
		filePath := filepath.Join(n.config.BasePath, subdir, fileName)
		if info, err := os.Stat(filePath); err == nil {
			return info, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", fileName)
}
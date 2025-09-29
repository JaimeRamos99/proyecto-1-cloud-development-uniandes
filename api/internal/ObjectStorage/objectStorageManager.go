package ObjectStorage

import (
	"proyecto1/root/internal/ObjectStorage/providers"
)

// FileStorageManager manages file storage operations using a provider
type FileStorageManager struct {
	provider providers.IFileStorageProvider
}

// NewFileStorageManager creates a new FileStorageManager with the given provider
func NewFileStorageManager(provider providers.IFileStorageProvider) *FileStorageManager {
	return &FileStorageManager{
		provider: provider,
	}
}

// UploadFile uploads a file using the configured provider
func (fsm *FileStorageManager) UploadFile(fileBuffer []byte, fileName string) error {
	return fsm.provider.UploadFile(fileBuffer, fileName)
}

// GetSignedUrl gets a signed URL for the file using the configured provider
func (fsm *FileStorageManager) GetSignedUrl(fileName string) (string, error) {
	return fsm.provider.GetSignedUrl(fileName)
}

// DeleteFile deletes a file using the configured provider
func (fsm *FileStorageManager) DeleteFile(fileName string) error {
	return fsm.provider.DeleteFile(fileName)
}

// Additional methods for NFS provider
func (fsm *FileStorageManager) MoveFile(srcFileName, destFileName, destSubdir string) error {
	if nfsProvider, ok := fsm.provider.(*providers.NFSProvider); ok {
		return nfsProvider.MoveFile(srcFileName, destFileName, destSubdir)
	}
	// For non-NFS providers, implement alternative logic or return error
	return fsm.provider.DeleteFile(srcFileName) // Simple fallback
}

func (fsm *FileStorageManager) GetFilePath(fileName, subdir string) string {
	if nfsProvider, ok := fsm.provider.(*providers.NFSProvider); ok {
		return nfsProvider.GetFilePath(fileName, subdir)
	}
	return fileName // Fallback for other providers
}

func (fsm *FileStorageManager) FileExists(fileName string) bool {
	if nfsProvider, ok := fsm.provider.(*providers.NFSProvider); ok {
		return nfsProvider.FileExists(fileName)
	}
	// For other providers, try to get signed URL as existence check
	_, err := fsm.provider.GetSignedUrl(fileName)
	return err == nil
}
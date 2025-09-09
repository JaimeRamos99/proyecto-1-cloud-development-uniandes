package ObjectStorage

import "worker/internal/ObjectStorage/providers"

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

// DownloadFile downloads a file using the configured provider
func (fsm *FileStorageManager) DownloadFile(fileName string) ([]byte, error) {
	return fsm.provider.DownloadFile(fileName)
}

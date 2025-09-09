# Object Storage - Strategy Pattern Implementation

Esta implementación utiliza el patrón Strategy para desacoplar completamente la lógica de negocio de los detalles de infraestructura de almacenamiento.

## Estructura

```
ObjectStorage/
├── objectStorageManager.go    # Manager que usa cualquier provider
├── example.go                 # Ejemplos de uso
├── README.md                  # Este archivo
└── providers/
    ├── s3Provider.go          # Implementación para AWS S3/LocalStack
    └── localProvider.go       # Implementación para almacenamiento local
```

## Uso Básico

### 1. Con S3/LocalStack

```go
import (
    "proyecto1/root/internal/ObjectStorage"
    "proyecto1/root/internal/ObjectStorage/providers"
)

// Configurar S3 provider
s3Config := &providers.S3Config{
    AccessKeyID:     "test",
    SecretAccessKey: "test",
    Region:          "us-east-1",
    BucketName:      "mi-bucket",
    EndpointURL:     "http://localhost:4566", // LocalStack
}

s3Provider, err := providers.NewS3Provider(s3Config)
if err != nil {
    log.Fatal(err)
}

// Crear manager
storageManager := ObjectStorage.NewFileStorageManager(s3Provider)

// Usar
fileContent := []byte("file content")
err = storageManager.UploadFile(fileContent, "videos/mi-video.mp4")
```

### 2. Con Almacenamiento Local

```go
// Configurar local provider
localConfig := &providers.LocalConfig{
    BasePath: "./storage",
}

localProvider, err := providers.NewLocalProvider(localConfig)
if err != nil {
    log.Fatal(err)
}

// Crear manager (misma interfaz!)
storageManager := ObjectStorage.NewFileStorageManager(localProvider)

// Usar exactamente igual
err = storageManager.UploadFile(fileContent, "videos/mi-video.mp4")
```

## Interfaz

El provider debe implementar `IFileStorageProvider`:

```go
type IFileStorageProvider interface {
    UploadFile(fileBuffer []byte, fileName string) error
    GetSignedUrl(fileName string) (string, error)
    DeleteFile(fileName string) error
}
```

## Ventajas

1. **Desacoplamiento total**: La lógica de negocio no conoce detalles de implementación
2. **Intercambiable**: Cambiar de S3 a local storage solo requiere cambiar el provider
3. **Testeable**: Fácil crear mocks para testing
4. **Extensible**: Agregar nuevos providers (GCS, Azure) es trivial
5. **Simple**: Solo 3 métodos esenciales

## Uso en Servicios de Negocio

```go
func (vs *VideoService) ProcessVideo(storageManager *ObjectStorage.FileStorageManager) {
    // Your logic doesn't know or care where the file is stored
    thumbnail := generateThumbnail()
    err := storageManager.UploadFile(thumbnail, "thumbnails/video123.jpg")

    videoURL, err := storageManager.GetSignedUrl("videos/video123.mp4")
    // Usar videoURL...
}
```

## Configuración

En desarrollo con LocalStack:

- Usar `S3Provider` con `EndpointURL: "http://localhost:4566"`

En desarrollo simple:

- Usar `LocalProvider` con `BasePath: "./storage"`

En producción:

- Usar `S3Provider` sin `EndpointURL` (usa AWS real)

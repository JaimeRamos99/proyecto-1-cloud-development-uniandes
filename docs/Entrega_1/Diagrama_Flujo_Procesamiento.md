# Diagrama de Flujo de Procesamiento de Videos

## Introducción

Este documento describe detalladamente el flujo completo de **carga, procesamiento y entrega** de archivos de video en el sistema. Desde la subida inicial por parte del usuario hasta la entrega del video procesado y listo para reproducción.

---

## Diagrama de Flujo General

```
┌─────────────────┐     ┌──────────────────┐     ┌───────────────────┐     ┌──────────────────┐
│                 │     │                  │     │                   │     │                  │
│     USUARIO     │────►│       API        │────►│      WORKER       │────►│    ENTREGA       │
│   (Frontend)    │     │   (Backend)      │     │   (Procesador)    │     │   (Video Listo)  │
│                 │     │                  │     │                   │     │                  │
└─────────────────┘     └──────────────────┘     └───────────────────┘     └──────────────────┘
```

---

## Flujo Detallado de Procesamiento

### Fase 1: Recepción y Validación (API)

```
    ┌─────────────────────────────────────────────────────────────────────────┐
    │                              FASE 1: API                               │
    └─────────────────────────────────────────────────────────────────────────┘

    [1] Usuario sube video
            │
            ▼
    [2] Validación de archivo
            │
    ┌───────────────────────┐
    │   FFprobe Validator   │
    │                       │
    │ • Formato válido      │
    │ • Duración ≤ 30s      │
    │ • Resolución válida   │
    │ • Tamaño ≤ 100MB      │
    │ • Integridad          │
    └─────────┬─────────────┘
            │ ✓ VÁLIDO
            ▼
    [3] Crear registro en BD
    ┌───────────────────────┐
    │     PostgreSQL        │
    │                       │
    │ INSERT videos {       │
    │   id: auto_increment  │
    │   title: "..."        │
    │   status: "uploaded"  │
    │   is_public: boolean  │
    │   user_id: 123        │
    │   uploaded_at: NOW()  │
    │ }                     │
    └─────────┬─────────────┘
            │
            ▼
    [4] Subir archivo a S3
    ┌───────────────────────┐
    │      S3 Storage       │
    │                       │
    │ Key: original/456.mp4 │
    │ Bucket: videos        │
    │ Size: file_size       │
    └─────────┬─────────────┘
            │
            ▼
    [5] Enviar mensaje SQS
    ┌───────────────────────┐
    │    SQS Message        │
    │                       │
    │ {                     │
    │   "s3_key":           │
    │   "original/456.mp4", │
    │   "video_id": 456,    │
    │   "timestamp": "..."  │
    │ }                     │
    └─────────┬─────────────┘
            │
            ▼
    [6] Respuesta al usuario
    ┌───────────────────────┐
    │    HTTP 201 Created   │
    │                       │
    │ {                     │
    │   "id": 456,          │
    │   "status": "uploaded"│
    │   "s3_key": "..."     │
    │ }                     │
    └───────────────────────┘
```

### Fase 2: Procesamiento Asíncrono (Worker)

```
    ┌─────────────────────────────────────────────────────────────────────────┐
    │                            FASE 2: WORKER                              │
    └─────────────────────────────────────────────────────────────────────────┘

    [7] Worker escucha SQS
    ┌───────────────────────┐
    │   Message Consumer    │
    │                       │
    │ • Poll SQS queue      │
    │ • Parse mensaje JSON  │
    │ • Extract video_id    │
    │ • Validate message    │
    └─────────┬─────────────┘
            │
            ▼
    [8] Verificar estado en BD
    ┌───────────────────────┐
    │    Database Check     │
    │                       │
    │ SELECT * FROM videos  │
    │ WHERE id = 456        │
    │                       │
    │ ✓ Status = "uploaded" │
    │ ✓ Video existe        │
    └─────────┬─────────────┘
            │
            ▼
    [9] Descargar video original
    ┌───────────────────────┐
    │   S3 Download         │
    │                       │
    │ GET original/456.mp4  │
    │ → videoData (bytes)   │
    └─────────┬─────────────┘
            │
            ▼
    [10] Procesar video
    ┌───────────────────────┐
    │   Video Processor     │
    │                       │
    │ • Crear archivo temp  │
    │ • Ejecutar FFmpeg     │
    │   - Resize 1280x720   │
    │   - Trim max 30s      │
    │   - Add watermark     │
    │   - Add intro/outro   │
    │   - Codec H.264       │
    │   - Quality CRF 23    │
    │ • Leer resultado      │
    │ • Cleanup temp files  │
    └─────────┬─────────────┘
            │
            ▼
    [11] Subir video procesado
    ┌───────────────────────┐
    │    S3 Upload          │
    │                       │
    │ PUT processed/456.mp4 │
    │ Content: processed    │
    │ video data            │
    └─────────┬─────────────┘
            │
            ▼
    [12] Actualizar estado BD
    ┌───────────────────────┐
    │   Database Update     │
    │                       │
    │ UPDATE videos SET     │
    │   status='processed', │
    │   processed_at=NOW()  │
    │ WHERE id=456          │
    └─────────┬─────────────┘
            │
            ▼
    [13] Eliminar mensaje SQS
    ┌───────────────────────┐
    │  Message Cleanup      │
    │                       │
    │ DELETE message from   │
    │ SQS queue             │
    │ (ProcessingComplete)  │
    └───────────────────────┘
```

### Fase 3: Entrega al Usuario (API)

```
    ┌─────────────────────────────────────────────────────────────────────────┐
    │                           FASE 3: ENTREGA                              │
    └─────────────────────────────────────────────────────────────────────────┘

    [14] Usuario solicita video
            │
            ▼
    [15] Verificar estado
    ┌───────────────────────┐
    │   Database Query      │
    │                       │
    │ SELECT status FROM    │
    │ videos WHERE id=456   │
    │                       │
    │ ✓ status="processed"  │
    └─────────┬─────────────┘
            │
            ▼
    [16] Generar URLs firmadas
    ┌───────────────────────┐
    │   Presigned URLs      │
    │                       │
    │ original_url:         │
    │   S3 signed URL       │
    │   (owner only)        │
    │                       │
    │ processed_url:        │
    │   S3 signed URL       │
    │   (public/owner)      │
    │                       │
    │ Expires: 1 hour       │
    └─────────┬─────────────┘
            │
            ▼
    [17] Respuesta al usuario
    ┌───────────────────────┐
    │     HTTP 200 OK       │
    │                       │
    │ {                     │
    │   "video_id": 456,    │
    │   "status":"processed"│
    │   "original_url":"...",│
    │   "processed_url":"..." │
    │   "votes": 0,         │
    │   "processed_at":"..."│
    │ }                     │
    └───────────────────────┘
```

---

## Manejo de Errores y Reintentos

### Estrategia de Reintentos

```
    ┌─────────────────────────────────────────────────────────────────────────┐
    │                          MANEJO DE ERRORES                             │
    └─────────────────────────────────────────────────────────────────────────┘

    Error en procesamiento
            │
            ▼
    ¿Es error permanente?
    ┌───────────────────────┐
    │   Error Classification│
    │                       │
    │ PERMANENTES:          │
    │ • Video no encontrado │
    │ • Ya procesado        │
    │ • Formato inválido    │
    │                       │
    │ TEMPORALES:           │
    │ • Error de red        │
    │ • S3 no disponible    │
    │ • DB timeout          │
    │ • FFmpeg crash        │
    └─────┬─────────┬───────┘
          │ SÍ      │ NO
          ▼         ▼
    [Descartar]   [Reintentar]
    mensaje       con backoff
          │         │
          ▼         ▼
    ┌─────────┐   ┌─────────────┐
    │  Dead   │   │ Exponential │
    │ Letter  │   │  Backoff:   │
    │ Queue   │   │             │
    │         │   │ Intento 1:  │
    │ (Para   │   │  2s delay   │
    │ análisis│   │ Intento 2:  │
    │ manual) │   │  4s delay   │
    │         │   │ Intento 3:  │
    │         │   │  8s delay   │
    │         │   │             │
    │         │   │ Max: 3      │
    │         │   │ intentos    │
    └─────────┘   └─────────────┘
```

### Configuración de Reintentos

| Parámetro             | Valor | Descripción                         |
| --------------------- | ----- | ----------------------------------- |
| **MaxRetries**        | 3     | Máximo número de reintentos         |
| **BaseDelay**         | 2s    | Delay base para exponential backoff |
| **MaxDelay**          | 16s   | Delay máximo entre intentos         |
| **EnableBackoff**     | true  | Activar estrategia de backoff       |
| **VisibilityTimeout** | 300s  | Tiempo de visibilidad en SQS        |

---

## Estados del Video

### Máquina de Estados

```
    [uploaded] ──────► [processed]
```

### Descripción de Estados

| Estado        | Descripción                                         | Acciones Permitidas                                                |
| ------------- | --------------------------------------------------- | ------------------------------------------------------------------ |
| **uploaded**  | Video subido y validado, pendiente de procesamiento | • Listar<br>• Ver detalles<br>• Eliminar (si privado)              |
| **processed** | Video procesado y listo para reproducción           | • Listar<br>• Ver detalles<br>• Reproducir<br>• Votar (si público) |

---

## Transformaciones de Video (FFmpeg)

### Pipeline de Procesamiento

```
    Input Video
         │
         ▼
    [1] Validación FFprobe
         │ ✓ Formato válido
         ▼
    [2] Recorte temporal
         │ Máximo 30 segundos
         ▼
    [3] Redimensión
         │ 1280x720 (720p 16:9)
         ▼
    [4] Agregar watermark
         │ Logo ANB en esquina
         ▼
    [5] Agregar intro/outro
         │ Bumpers corporativos
         ▼
    [6] Codificación H.264
         │ CRF 23 (alta calidad)
         ▼
    [7] Output Video
         │ .mp4 optimizado
         ▼
    Processed Video
```

### Comando FFmpeg Generado

```bash
ffmpeg -i /tmp/input_video.mp4 \
  -i /app/assets/watermark.png \
  -filter_complex "[0:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:-1:-1:color=black[scaled];[scaled][1:v]overlay=W-w-10:10[watermarked]" \
  -map "[watermarked]" \
  -map 0:a? \
  -c:v libx264 \
  -crf 23 \
  -preset medium \
  -c:a aac \
  -b:a 128k \
  -movflags +faststart \
  -t 30 \
  /tmp/output_video.mp4
```

---

## Almacenamiento S3

### Estructura de Archivos

```
proyecto1-videos/
├── original/
│   ├── 1.mp4
│   ├── 2.mp4
│   └── 456.mp4
└── processed/
    ├── 1.mp4
    ├── 2.mp4
    └── 456.mp4
```

### Políticas de Acceso

| Directorio     | Acceso                    | Descripción                                |
| -------------- | ------------------------- | ------------------------------------------ |
| **original/**  | Propietario únicamente    | Videos originales sin procesar             |
| **processed/** | Público (con URL firmada) | Videos procesados listos para reproducción |

### URLs Presignadas

- **Duración**: 1 hora
- **Permissions**: Read-only
- **Uso**: Streaming y descarga temporal
- **Seguridad**: No exposición de credenciales

---

## Monitoreo y Métricas

### Métricas Clave

| Métrica                 | Descripción                      | Objetivo    |
| ----------------------- | -------------------------------- | ----------- |
| **Upload Success Rate** | % de uploads exitosos            | > 99%       |
| **Processing Time**     | Tiempo promedio de procesamiento | < 60s       |
| **Queue Depth**         | Mensajes pendientes en SQS       | < 10        |
| **Error Rate**          | % de errores en procesamiento    | < 1%        |
| **Storage Usage**       | Uso de almacenamiento S3         | Monitoreado |

### Logs Estructurados

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "component": "worker",
  "video_id": 456,
  "s3_key": "original/456.mp4",
  "action": "processing_started",
  "duration_ms": 0,
  "file_size_bytes": 5242880
}
```

---

## Consideraciones de Rendimiento

### Optimizaciones Implementadas

1. **Streaming Upload**: Procesamiento en chunks para archivos grandes
2. **Parallel Processing**: Worker pool para procesamiento concurrente
3. **Efficient Storage**: Separación de archivos originales y procesados
4. **Caching Strategy**: URLs presignadas con TTL optimizado
5. **Resource Cleanup**: Eliminación automática de archivos temporales

### Escalabilidad

- **Horizontal**: Múltiples instancias de worker
- **Vertical**: Ajuste de recursos FFmpeg según carga
- **Storage**: Auto-scaling de S3 según demanda
- **Queue**: SQS maneja picos de carga automáticamente

---

## Seguridad

### Validaciones de Seguridad

1. **File Validation**: FFprobe valida integridad y formato
2. **Size Limits**: 100MB máximo por archivo
3. **Duration Limits**: 30 segundos máximo
4. **Format Restrictions**: Solo formatos de video permitidos
5. **Access Control**: URLs firmadas con expiración

### Protección contra Ataques

- **Malicious Files**: Validación completa con FFprobe
- **Resource Exhaustion**: Límites de tiempo y memoria
- **Unauthorized Access**: JWT tokens y URLs firmadas
- **Injection Attacks**: Sanitización de parámetros FFmpeg

---

## Recuperación ante Fallos

### Estrategias de Recuperación

| Fallo              | Estrategia                               | Tiempo Recuperación |
| ------------------ | ---------------------------------------- | ------------------- |
| **Worker Down**    | Auto-restart + health checks             | 30s                 |
| **S3 Unavailable** | Retry con backoff exponencial            | 5min                |
| **Database Down**  | Connection pooling + retry               | 10s                 |
| **FFmpeg Error**   | Reintentar con configuración alternativa | 1min                |
| **SQS Issues**     | Dead Letter Queue + manual review        | Manual              |

### Backup y Restauración

- **Database**: Backups automáticos cada 4 horas
- **S3 Objects**: Versionado habilitado
- **Configuration**: Infrastructure as Code
- **Monitoring**: Alertas automáticas de fallos

---

## Resumen del Flujo

### Tiempo Total del Proceso

| Fase                 | Tiempo Estimado     | Notas                       |
| -------------------- | ------------------- | --------------------------- |
| **Validación**       | 2-5 segundos        | Depende del tamaño          |
| **Upload S3**        | 5-30 segundos       | Depende de conexión         |
| **Queue Processing** | < 1 segundo         | Casi instantáneo            |
| **Video Processing** | 10-60 segundos      | Depende de duración         |
| **Final Upload**     | 5-20 segundos       | Video procesado más pequeño |
| **DB Update**        | < 1 segundo         | Operación rápida            |
| **TOTAL**            | **22-116 segundos** | Para video de 30s, 100MB    |

### Puntos de Fallo y Mitigation

1. ✅ **Validación falla** → Error inmediato al usuario
2. ✅ **Upload falla** → Retry automático con backoff
3. ✅ **Processing falla** → Retry hasta 3 veces
4. ✅ **Worker crash** → Mensaje regresa a queue
5. ✅ **Final upload falla** → Retry con exponential backoff

**¡Sistema robusto con alta disponibilidad y recuperación automática!** 🚀

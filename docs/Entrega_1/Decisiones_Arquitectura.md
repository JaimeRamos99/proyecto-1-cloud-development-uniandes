# Decisiones de Arquitectura

## Introducción

Este documento describe las decisiones arquitectónicas clave tomadas durante el desarrollo del sistema de gestión de videos, explicando el razonamiento detrás de cada elección.

---

## 🏗️ **Arquitectura General**

### **Microservicios con Comunicación Asíncrona**

**Decisión**: Separar la API REST del Worker de procesamiento usando SQS para comunicación.

**Razón**:

- Permite escalabilidad independiente de cada servicio
- Desacopla el procesamiento pesado de la respuesta HTTP
- Facilita el mantenimiento y deployment independiente

**Implementación**: API envía mensajes SQS, Worker procesa asincrónicamente.

---

## 🔄 **Procesamiento de Videos**

### **Buffer Completo en Memoria**

**Decisión**: Cargar todo el archivo de video en memoria como buffer antes del procesamiento.

**Razón**:

- Simplifica la implementación para los requisitos del proyecto
- FFmpeg requiere acceso completo al archivo para procesamiento
- Los videos están limitados a 100MB (tamaño manejable en memoria)

**Alternativa considerada**: Streaming processing (más complejo, no necesario para el alcance).

### **Prefijos de Estado en S3 Keys**

**Decisión**: Usar prefijos que representan estados de procesamiento: `original/` y `processed/`.

**Razón**:

- Claridad en la organización de archivos
- Facilita debugging y monitoreo
- Permite diferentes políticas de acceso por estado
- Mantiene el archivo original como backup

**Estructura**:

```
proyecto1-videos/
├── original/123.mp4    # Video sin procesar
└── processed/123.mp4   # Video procesado
```

---

## 📊 **Base de Datos y Rankings**

### **Vista Materializada para Rankings**

**Decisión**: Usar una vista materializada de PostgreSQL para rankings dinámicos.

**Razón**:

- Performance optimizada para consultas complejas de ranking
- Actualización automática cada minuto
- Evita cálculos costosos en tiempo real
- Soporte nativo de PostgreSQL para este patrón

**Implementación**:

```sql
CREATE MATERIALIZED VIEW player_rankings AS
SELECT u.*, COUNT(v.id) as total_votes,
       ROW_NUMBER() OVER (ORDER BY COUNT(v.id) DESC) as ranking
FROM users u LEFT JOIN videos v ON u.id = v.user_id
GROUP BY u.id;
```

### **Soft Delete para Videos**

**Decisión**: Usar `deleted_at` timestamp en lugar de eliminación física.

**Razón**:

- Preserva integridad referencial en votos
- Permite recuperación de datos
- Mantiene auditoría de eliminaciones
- Videos públicos no se pueden eliminar (integridad de rankings)

---

## 🔄 **Manejo de Errores y Resilencia**

### **Exponential Backoff en Worker**

**Decisión**: Implementar retry con backoff exponencial para errores temporales.

**Razón**:

- Reduce carga en servicios que están fallando
- Aumenta probabilidad de éxito en recuperación
- Evita "thundering herd" en reintentos

**Configuración**:

- Max retries: 3
- Base delay: 2s
- Max delay: 16s
- Errores permanentes no se reintentan

### **Long Polling en SQS**

**Decisión**: Usar long polling (20s) en lugar de polling corto.

**Razón**:

- Reduce costos de AWS SQS (menos requests)
- Menor latencia en recepción de mensajes
- Uso eficiente de recursos del worker

---

## ☁️ **Almacenamiento y Cloud**

### **AWS SDK con LocalStack para Desarrollo**

**Decisión**: Usar AWS SDK oficial con LocalStack para desarrollo local.

**Razón**:

- Consistencia entre desarrollo y producción
- Testing realista de integraciones AWS
- Facilita migración a AWS real
- LocalStack simula fielmente S3 y SQS

### **URLs Presignadas para Acceso Seguro**

**Decisión**: Generar URLs presignadas en lugar de acceso directo a S3.

**Razón**:

- Seguridad: no exponer credenciales AWS
- Control de acceso temporal (1 hora)
- Flexibilidad en políticas de acceso
- Compatibilidad con CDNs

---

## 🐳 **Containerización y Orquestación**

### **Docker Compose Multi-Servicio**

**Decisión**: Usar docker-compose para orquestar todos los servicios localmente.

**Razón**:

- Simplicidad en desarrollo local
- Consistencia de entorno entre desarrolladores
- Fácil setup con un solo comando
- Health checks para dependencias

**Servicios incluidos**:

- **PostgreSQL**: Base de datos principal
- **Nginx**: Proxy reverso y load balancer
- **API**: Servicio REST en Go
- **Worker**: Procesador de videos
- **LocalStack**: Simulador AWS (S3, SQS)
- **Swagger UI**: Documentación interactiva

### **Health Checks y Dependencias**

**Decisión**: Implementar health checks y dependencias entre servicios.

**Razón**:

- Asegura orden correcto de startup
- Evita errores por servicios no listos
- Facilita debugging de problemas de conectividad

---

## 🔐 **Seguridad y Autenticación**

### **JWT Tokens para Autenticación**

**Decisión**: Usar JWT en lugar de sesiones server-side.

**Razón**:

- Stateless: no requiere almacenamiento de sesión
- Escalabilidad: funciona con múltiples instancias
- Estándar de la industria para APIs REST
- Fácil integración con frontend

### **Rate Limiting por Nginx**

**Decisión**: Implementar rate limiting a nivel de proxy reverso.

**Razón**:

- Protección contra abuso y DDoS
- Diferentes límites por tipo de endpoint
- Eficiencia: bloquea requests antes de llegar a la API
- Configuración centralizada

---

## 📈 **Monitoreo y Observabilidad**

### **Logs Estructurados**

**Decisión**: Usar logs estructurados en formato JSON.

**Razón**:

- Facilita parsing y análisis automatizado
- Compatibilidad con sistemas de logging (ELK, Fluentd)
- Debugging más eficiente
- Métricas extraíbles automáticamente

### **Health Endpoints**

**Decisión**: Implementar endpoints de health check para todos los servicios.

**Razón**:

- Monitoreo de disponibilidad
- Detección temprana de problemas
- Integración con orquestadores (Docker, Kubernetes)
- Load balancer health checks

---

## 🎯 **Decisiones de Diseño de API**

### **RESTful con Recursos Anidados**

**Decisión**: Diseñar API REST con recursos anidados lógicos.

**Razón**:

- Intuitivo para desarrolladores
- Estándar de la industria
- Fácil documentación con OpenAPI/Swagger
- Separación clara de responsabilidades

**Ejemplos**:

- `POST /api/videos/upload` - Upload de video
- `POST /api/public/videos/{id}/vote` - Votar video
- `GET /api/public/rankings` - Ver rankings

### **Validación con FFprobe**

**Decisión**: Usar FFprobe para validación de archivos de video.

**Razón**:

- Validación robusta de formato y contenido
- Detección de archivos corruptos
- Información detallada de metadatos
- Herramienta estándar de la industria

---

## 🚀 **Escalabilidad y Performance**

### **Índices de Base de Datos Optimizados**

**Decisión**: Crear índices específicos para consultas frecuentes.

**Razón**:

- Mejora performance de queries
- Soporte para filtros complejos en rankings
- Optimización para paginación
- Reducción de tiempo de respuesta

**Índices clave**:

- `idx_videos_user_id_is_public` - Videos por usuario y visibilidad
- `idx_player_rankings_total_votes` - Rankings por votos
- `idx_votes_user_video` - Votos únicos por usuario-video

---

## 📋 **Resumen de Beneficios**

| Decisión                 | Beneficio Principal               |
| ------------------------ | --------------------------------- |
| **Microservicios**       | Escalabilidad independiente       |
| **SQS + Worker**         | Procesamiento asíncrono confiable |
| **Vista Materializada**  | Rankings rápidos y actualizados   |
| **Docker Compose**       | Desarrollo local simplificado     |
| **JWT + Nginx**          | Seguridad y performance           |
| **AWS SDK + LocalStack** | Consistencia dev/prod             |

---

**Estas decisiones arquitectónicas proporcionan un sistema robusto, escalable y mantenible que cumple con todos los requisitos del proyecto.** 🚀

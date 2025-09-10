# Diagramas C4 - Sistema de Gestión de Videos

## Introducción

Este documento presenta la arquitectura del sistema de gestión de videos utilizando el modelo C4 (Context, Containers, Components, Code). El sistema permite a los usuarios subir videos, procesarlos, votarlos y generar rankings basados en popularidad.

---

## Nivel 1 - Diagrama de Contexto

### Descripción

Muestra el sistema de gestión de videos y cómo interactúa con usuarios y sistemas externos.

```
                    ┌─────────────────────────┐
                    │                         │
                    │       USUARIOS          │
                    │   (Personas Físicas)    │
                    │                         │
                    └───────────┬─────────────┘
                                │
                                │ Sube videos, vota,
                                │ consulta rankings
                                ▼
            ┌───────────────────────────────────────────────┐
            │                                               │
            │        SISTEMA DE GESTIÓN DE VIDEOS          │
            │                                               │
            │  • Permite subir y gestionar videos          │
            │  • Sistema de votación                       │
            │  • Generación de rankings                    │
            │  • Procesamiento automático de videos       │
            │                                               │
            └─────────────┬─────────────────────────────────┘
                          │
                          │ Almacena videos y datos
                          ▼
                ┌─────────────────────────┐
                │                         │
                │   SERVICIOS AWS/CLOUD   │
                │                         │
                │  • S3 (Almacenamiento)  │
                │  • SQS (Mensajería)     │
                │                         │
                └─────────────────────────┘
```

### Elementos del Contexto

| Actor/Sistema                    | Tipo                | Descripción                                                                                   |
| -------------------------------- | ------------------- | --------------------------------------------------------------------------------------------- |
| **Usuarios**                     | Persona             | Usuarios finales que interactúan con el sistema para subir videos, votar y consultar rankings |
| **Sistema de Gestión de Videos** | Sistema de Software | Aplicación principal que gestiona videos, usuarios, votos y rankings                          |
| **Servicios AWS/Cloud**          | Sistema Externo     | Servicios de almacenamiento (S3) y mensajería (SQS) para el procesamiento asíncrono           |

---

## Nivel 2 - Diagrama de Contenedores

### Descripción

Muestra los contenedores principales que componen el sistema y sus interacciones.

```
┌─────────────┐          ┌─────────────────────────────────────────────┐
│             │          │                                             │
│  USUARIOS   │          │           SISTEMA DE VIDEOS                │
│   (Web)     │          │                                             │
│             │          │  ┌─────────────┐    ┌──────────────────┐   │
└─────┬───────┘          │  │             │    │                  │   │
      │                  │  │  FRONTEND   │    │      NGINX       │   │
      │ HTTPS            │  │  (React)    │◄───┤  (Reverse Proxy/ │   │
      └──────────────────┼──► Puerto 3000 │    │  Load Balancer)  │   │
                         │  │             │    │   Puerto 80      │   │
                         │  └─────────────┘    └─────────┬────────┘   │
                         │                                │            │
                         │  ┌─────────────┐               │ HTTP       │
                         │  │             │◄──────────────┘            │
                         │  │  API REST   │                             │
                         │  │    (Go)     │                             │
                         │  │ Puerto 8080 │                             │
                         │  │             │                             │
                         │  └─────┬───────┘                             │
                         │        │                                     │
                         │        │ SQL                                 │
                         │        ▼                                     │
                         │  ┌─────────────┐    ┌──────────────────┐   │
                         │  │             │    │                  │   │
                         │  │ PostgreSQL  │    │     WORKER       │   │
                         │  │  Database   │◄───┤  (Procesamiento  │   │
                         │  │ Puerto 5432 │    │   de Videos)     │   │
                         │  │             │    │      (Go)        │   │
                         │  └─────────────┘    └──────┬───────────┘   │
                         │                             │               │
                         │  ┌─────────────┐            │               │
                         │  │             │            │ SQS           │
                         │  │ SWAGGER UI  │            │               │
                         │  │   (Docs)    │            ▼               │
                         │  │ Puerto 8081 │    ┌──────────────────┐   │
                         │  │             │    │                  │   │
                         │  └─────────────┘    │   LOCALSTACK     │   │
                         │                      │                  │   │
                         └──────────────────────┤ • S3 (Videos)    │   │
                                                │ • SQS (Mensajes) │   │
                                                │ Puerto 4566      │   │
                                                │                  │   │
                                                └──────────────────┘   │
                                                                       │
                                        ┌──────────────────────────────┘
                                        │
                                        ▼
                              ┌─────────────────────────┐
                              │                         │
                              │   SERVICIOS EXTERNOS    │
                              │                         │
                              │ • AWS S3 (Producción)   │
                              │ • AWS SQS (Producción)  │
                              │                         │
                              └─────────────────────────┘
```

### Contenedores del Sistema

| Contenedor     | Tecnología    | Puerto | Descripción                   | Responsabilidades                                                                         |
| -------------- | ------------- | ------ | ----------------------------- | ----------------------------------------------------------------------------------------- |
| **Frontend**   | React         | 3000   | Interfaz web de usuario       | • Interfaz para subir videos<br>• Sistema de votación<br>• Visualización de rankings      |
| **Nginx**      | Nginx         | 80     | Proxy reverso y load balancer | • Enrutamiento de requests<br>• Rate limiting<br>• CORS handling<br>• SSL termination     |
| **API REST**   | Go/Gin        | 8080   | API backend principal         | • Gestión de usuarios<br>• CRUD de videos<br>• Sistema de votación<br>• Autenticación JWT |
| **Worker**     | Go            | -      | Procesador de videos          | • Procesa videos asincrónicamente<br>• Genera thumbnails<br>• Convierte formatos          |
| **PostgreSQL** | PostgreSQL 15 | 5432   | Base de datos principal       | • Almacena usuarios<br>• Metadata de videos<br>• Votos y rankings                         |
| **LocalStack** | LocalStack    | 4566   | Simulador AWS (desarrollo)    | • S3 para almacenar videos<br>• SQS para cola de mensajes                                 |
| **Swagger UI** | Swagger       | 8081   | Documentación API             | • Documentación interactiva<br>• Testing de endpoints                                     |

### Flujos de Comunicación

1. **Usuario → Frontend → Nginx → API**: Operaciones CRUD y autenticación
2. **API → PostgreSQL**: Persistencia de datos
3. **API → LocalStack/S3**: Almacenamiento de videos
4. **API → LocalStack/SQS**: Envío de mensajes de procesamiento
5. **Worker → LocalStack/SQS**: Consumo de mensajes
6. **Worker → PostgreSQL**: Actualización del estado de procesamiento

---

## Nivel 3 - Diagrama de Componentes (API)

### Descripción

Detalle de los componentes internos de la API REST.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               API REST (Go/Gin)                                │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐ │
│  │                     │    │                      │    │                     │ │
│  │   HTTP HANDLERS     │    │     MIDDLEWARES      │    │      ROUTERS        │ │
│  │                     │    │                      │    │                     │ │
│  │ • AuthHandler       │◄───┤ • AuthMiddleware     │◄───┤ • /api/auth/*       │ │
│  │ • VideosHandler     │    │ • CORSMiddleware     │    │ • /api/videos/*     │ │
│  │ • VotesHandler      │    │ • LoggingMiddleware  │    │ • /api/votes/*      │ │
│  │ • RankingsHandler   │    │ • RateLimitMiddleware│    │ • /api/rankings/*   │ │
│  │ • HealthHandler     │    │                      │    │ • /api/health       │ │
│  │                     │    │                      │    │                     │ │
│  └──────────┬──────────┘    └──────────────────────┘    └─────────────────────┘ │
│             │                                                                   │
│             ▼                                                                   │
│  ┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐ │
│  │                     │    │                      │    │                     │ │
│  │       DTOs          │    │      SERVICES        │    │       AUTH          │ │
│  │  (Data Transfer)    │    │   (Business Logic)   │    │                     │ │
│  │                     │    │                      │    │ • JWT Manager       │ │
│  │ • UserDTO           │◄───┤ • UserService        │◄───┤ • Password Hash     │ │
│  │ • VideoDTO          │    │ • VideoService       │    │ • Session Store     │ │
│  │ • VoteDTO           │    │ • VoteService        │    │                     │ │
│  │ • RankingDTO        │    │ • RankingService     │    │                     │ │
│  │                     │    │                      │    │                     │ │
│  └─────────────────────┘    └──────────┬───────────┘    └─────────────────────┘ │
│                                         │                                       │
│                                         ▼                                       │
│  ┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐ │
│  │                     │    │                      │    │                     │ │
│  │     MODELS          │    │    REPOSITORIES      │    │   EXTERNAL APIS     │ │
│  │   (Data Models)     │    │   (Data Access)      │    │                     │ │
│  │                     │    │                      │    │ • S3 Provider       │ │
│  │ • User              │◄───┤ • UserRepository     │◄───┤ • SQS Provider      │ │
│  │ • Video             │    │ • VideoRepository    │    │ • Object Storage    │ │
│  │ • Vote              │    │ • VoteRepository     │    │   Manager           │ │
│  │ • PlayerRanking     │    │ • RankingRepository  │    │ • Messaging         │ │
│  │                     │    │                      │    │   Interface         │ │
│  └─────────────────────┘    └──────────┬───────────┘    └─────────────────────┘ │
│                                         │                                       │
│                                         ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                              DATABASE CONNECTION                            │ │
│  │                                PostgreSQL Driver                           │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Componentes de la API

| Capa                  | Componentes                    | Responsabilidades                                                                                       |
| --------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------- |
| **HTTP Layer**        | Handlers, Middlewares, Routers | • Manejo de requests HTTP<br>• Validación de entrada<br>• Autenticación/Autorización<br>• Rate limiting |
| **Business Layer**    | Services, DTOs                 | • Lógica de negocio<br>• Transformación de datos<br>• Validación de reglas de negocio                   |
| **Data Layer**        | Models, Repositories           | • Modelos de datos<br>• Acceso a base de datos<br>• Queries y operaciones CRUD                          |
| **Integration Layer** | External APIs                  | • Integración con S3<br>• Integración con SQS<br>• Gestión de almacenamiento                            |

---

## Nivel 3 - Diagrama de Componentes (Worker)

### Descripción

Detalle de los componentes internos del Worker de procesamiento de videos.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            WORKER (Procesamiento Videos)                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐ │
│  │                     │    │                      │    │                     │ │
│  │   MESSAGE CONSUMER  │    │    VIDEO PROCESSOR   │    │    FILE MANAGER     │ │
│  │                     │    │                      │    │                     │ │
│  │ • SQS Listener      │───►│ • Format Converter   │◄───┤ • S3 Download       │ │
│  │ • Message Parser    │    │ • Thumbnail Generator│    │ • S3 Upload         │ │
│  │ • Error Handler     │    │ • Quality Optimizer  │    │ • Temp Storage      │ │
│  │ • Retry Logic       │    │ • Metadata Extractor │    │ • File Validation   │ │
│  │                     │    │                      │    │                     │ │
│  └──────────┬──────────┘    └──────────┬───────────┘    └─────────────────────┘ │
│             │                          │                                       │
│             ▼                          ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                            PROCESSING PIPELINE                             │ │
│  ├─────────────────────────────────────────────────────────────────────────────┤ │
│  │                                                                             │ │
│  │  1. Download Video    2. Validate       3. Process        4. Upload        │ │
│  │     from S3         ───► Format      ───► Video        ───► Results        │ │
│  │                                                                             │ │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   ┌─────────────┐  │ │
│  │  │   FFPROBE   │    │   FFMPEG    │    │ THUMBNAIL   │   │ UPDATE DB   │  │ │
│  │  │ Validation  │    │ Conversion  │    │ Generator   │   │   Status    │  │ │
│  │  └─────────────┘    └─────────────┘    └─────────────┘   └─────────────┘  │ │
│  │                                                                             │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                                         │                                       │
│                                         ▼                                       │
│  ┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐ │
│  │                     │    │                      │    │                     │ │
│  │     DATABASE        │    │     MONITORING       │    │   ERROR HANDLING    │ │
│  │    CONNECTION       │    │                      │    │                     │ │
│  │                     │    │ • Processing Stats   │    │ • Dead Letter Queue │ │
│  │ • Video Repository  │◄───┤ • Performance Metrics│    │ • Retry Mechanism   │ │
│  │ • Status Updates    │    │ • Error Tracking     │◄───┤ • Error Logging     │ │
│  │                     │    │ • Health Checks      │    │ • Alerting          │ │
│  │                     │    │                      │    │                     │ │
│  └─────────────────────┘    └──────────────────────┘    └─────────────────────┘ │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Componentes del Worker

| Componente              | Descripción                              | Tecnologías               |
| ----------------------- | ---------------------------------------- | ------------------------- |
| **Message Consumer**    | Consume mensajes de SQS y maneja errores | Go, AWS SQS SDK           |
| **Video Processor**     | Procesa videos usando FFmpeg             | FFmpeg, FFprobe           |
| **File Manager**        | Gestiona descarga/subida de archivos     | AWS S3 SDK                |
| **Processing Pipeline** | Orquesta el flujo de procesamiento       | Go channels, workers pool |
| **Database Connection** | Actualiza estado en PostgreSQL           | PostgreSQL driver         |
| **Monitoring**          | Métricas y monitoreo del procesamiento   | Logs, metrics             |
| **Error Handling**      | Manejo de errores y reintentos           | DLQ, retry logic          |

---

## Patrones Arquitectónicos Implementados

### 1. **Microservicios**

- Separación clara entre API y Worker
- Comunicación asíncrona via SQS
- Escalabilidad independiente

### 2. **Patrón Repository**

- Separación entre lógica de negocio y acceso a datos
- Interfaces bien definidas para persistencia
- Facilita testing y cambios de BD

### 3. **Patrón MVC/Clean Architecture**

- Handlers (Controllers)
- Services (Use Cases/Business Logic)
- Repositories (Data Access)
- Models (Entities)

### 4. **Message Queue Pattern**

- Procesamiento asíncrono de videos
- Desacoplamiento entre componentes
- Resilencia con Dead Letter Queues

### 5. **Proxy Pattern**

- Nginx como proxy reverso
- Load balancing y SSL termination
- Rate limiting y CORS handling

---

## Consideraciones de Escalabilidad

### Horizontal Scaling

- **API**: Múltiples instancias detrás de load balancer
- **Worker**: Pool de workers escalable
- **Database**: Read replicas para consultas

### Vertical Scaling

- **Optimización de queries** con índices apropiados
- **Connection pooling** para base de datos
- **Caching** con Redis (futuro)

### Monitoreo y Observabilidad

- **Logs estructurados** en todos los componentes
- **Health checks** para todos los servicios
- **Métricas de performance** y errores
- **Tracing distribuido** (futuro con Jaeger/Zipkin)

---

## Seguridad

### Autenticación y Autorización

- **JWT tokens** para autenticación
- **Password hashing** con bcrypt
- **Session management** seguro

### Protección de APIs

- **Rate limiting** por IP y endpoint
- **CORS** configurado correctamente
- **Input validation** en todos los endpoints

### Almacenamiento Seguro

- **Signed URLs** para acceso a S3
- **Encryption at rest** en base de datos
- **Secrets management** con variables de entorno

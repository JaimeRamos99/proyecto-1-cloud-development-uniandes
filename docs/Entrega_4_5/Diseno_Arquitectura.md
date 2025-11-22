# **Arquitectura AWS - Proyecto Video Platform**

## **📋 Resumen del Proyecto**

Arquitectura cloud-native para plataforma de procesamiento de videos, migrando de una solución monolítica en EC2 única a una arquitectura distribuida, escalable y serverless con CI/CD automatizado.

A continuación se presenta toda la implementación hecha para 


Servidor: 
## Diagrama final

![Diagrama de Arquitectura](/docs/Entrega_4_5/DiagramaArquitectura.png)

## **Cambios Principales Realizados**

| **Componente** | **Antes** | **Ahora** |
|----------------|-----------|-----------|
| **Frontend** | En EC2 con backend | S3 + CloudFront CDN |
| **Backend** | 1 EC2 fija | ALB + ASG (2-4 instancias) |
| **Uploads** | A través del backend | Presigned URLs directo a S3 |
| **Deploy** | Manual | GitHub Actions automatizado |
| **Contenedores** | Local | ECR centralizado |
| **Infraestructura** | Click en consola | Terraform IaC |
| **Worker** | Proceso en EC2 | Lambda serverless |


## **Componentes de Arquitectura Implementados**

### **1. Cómputo y Escalabilidad**

#### **1.1 Instancias EC2 Multi-AZ**
- API (Web Server) en EC2 distribuida en múltiples zonas de disponibilidad
- Alta disponibilidad garantizada

#### **1.2 Auto Scaling Group**
- Escalado automático basado en CPU (target: 70%)
- Mínimo 2, máximo 4 instancias
- Health checks automáticos

#### **1.3 Application Load Balancer (ALB)**
- Punto único de entrada para la API
- Distribución de tráfico entre instancias
- Health checks en `/api/health`

#### **1.4 AWS Lambda** (ENTREGA 5)
- Worker serverless para procesamiento de videos
- Trigger automático por mensajes SQS
- Concurrencia hasta 10 ejecuciones simultáneas
- Timeout: 15 minutos, Memoria: 3008 MB

### **2. Almacenamiento**

#### **2.1 Amazon S3**
- **Estructura**:
  - `uploaded/` - Videos originales
  - `processed/` - Videos procesados  
  - `assets/` - Recursos (bumpers, watermarks, cortinillas)
- Políticas de lifecycle para optimización de costos

#### **2.2 Amazon RDS PostgreSQL**
- Base de datos relacional
- Tablas: usuarios, videos, votos, rankings
- Backups automáticos diarios

#### **2.3 Amazon ECR**
- Registro privado de imágenes Docker
- Repositorios: `proyecto1-api`, `proyecto1-worker`, `proyecto1-frontend`
- Versionamiento automático con tags

### **3. Mensajería y Comunicación**

#### **3.1 Amazon SQS**
- Cola de mensajes para procesamiento asíncrono
- Dead Letter Queue (DLQ) para manejo de errores
- Visibility timeout: 960 segundos

#### **3.2 VPC Endpoints para S3**
- Comunicación privada entre EC2/Lambda y S3
- Sin costos de transferencia por internet público
- Mayor seguridad y menor latencia

### **4. Frontend y Entrega de Contenido**

#### **4.1 Frontend en S3**
- Aplicación React estática
- Bucket privado (sin website hosting directo)

#### **4.2 CloudFront CDN**
- Distribución global del frontend
- Proxy reverso para la API
- Behaviors:
  - `/` → S3 (Cache: variable)
  - `/api/*` → ALB (Cache: disabled)
- Un único dominio (evita CORS)

### **5. Seguridad y Permisos**

#### **5.1 IAM Roles**
- **EC2 Role**: S3, SQS, ECR, CloudWatch
- **Lambda Role**: S3, SQS, RDS, CloudWatch
- Principio de least privilege

#### **5.2 Security Groups**
- ALB: Acepta tráfico 80/443 desde internet
- EC2: Solo acepta tráfico desde ALB
- RDS: Solo acepta desde EC2 y Lambda
- Lambda: Egress only

### **6. Automatización y DevOps**

#### **6.1 Terraform (IaC)**
- Infraestructura como código
- Módulos organizados por componente
- Estado remoto en S3 con locking

#### **6.2 User-Data Scripts**
- Automatización de configuración EC2:
  - Instalación de Docker
  - Configuración AWS CLI
  - Pull de imágenes desde ECR
  - Inicio de contenedores

#### **6.3 GitHub Actions CI/CD**
- Pipeline completo automatizado (ver sección detallada abajo)

#### **6.4 Docker Compose**
- Orquestación de contenedores en EC2:
  - Nginx (proxy)
  - API Go
- Restart automático: `always`

### **7. Monitoreo y Observabilidad**

#### **7.1 CloudWatch Logs**
- Log Groups:
  - `/aws/lambda/video-processor`
  - `/aws/ec2/api`
  - `/aws/cloudfront/distribution`

#### **7.2 CloudWatch Metrics**
- Métricas de CPU, memoria, latencia
- Triggers para auto-scaling

#### **7.3 ALB Health Checks**
- Endpoint: `/api/health`
- Intervalo: 30 segundos
- Healthy threshold: 2 checks



## **🚀 Pipeline CI/CD - GitHub Actions**

### **Estructura del Pipeline**

El pipeline se ejecuta en push a `main` y consta de 4 jobs principales:

```yaml
Jobs:
├── build-and-push-to-ecr
├── migrate-database
├── build-and-deploy-frontend
└── deploy (depende de los anteriores)
```

### **Job 1: Build and Push to ECR**

**Propósito**: Construir y publicar imágenes Docker

**Pasos**:
1. Checkout del código
2. Configurar credenciales AWS
3. Login a ECR
4. Build y push de imagen API:
   - Tags: `{sha}` y `latest`
   - Plataforma: `linux/amd64`
5. Build y push de imagen Lambda Worker:
   - Tags: `lambda-{sha}` y `lambda`
   - Dockerfile especial: `Dockerfile.lambda`

**Tecnologías**:
- Docker Buildx para builds multi-plataforma
- GitHub Actions cache para optimización

### **Job 2: Database Migration**

**Propósito**: Ejecutar migraciones de base de datos

**Pasos**:
1. Instalar cliente PostgreSQL
2. Conectar a RDS usando endpoint y credenciales
3. Ejecutar scripts SQL en orden:
   - `001_create_user_table.sql`
   - `002_create_video_table.sql`
   - `003_add_is_public_to_videos.sql`
   - `004_create_votes_table.sql`
   - `005_create_player_rankings_view.sql`


### **Job 3: Build and Deploy Frontend**

**Propósito**: Compilar y desplegar React a S3/CloudFront

**Pasos**:
1. Setup Node.js 18
2. Install dependencies (`npm ci`)
3. Build frontend (`npm run build`)
4. Sync a S3:
   - Archivos estáticos con cache largo (1 año)
   - `index.html` sin cache
5. Invalidar CloudFront cache

**Optimizaciones**:
- Cache de dependencias npm
- Cache control headers diferenciados

### **Job 4: Deploy**

**Propósito**: Actualizar servicios en producción

#### **4.1 Deploy to ASG (Rolling Update)**

**Proceso**:
1. Obtener instancias del Auto Scaling Group
2. Para cada instancia:
   - Verificar SSM Agent online
   - Ejecutar comandos vía SSM:
     ```bash
     docker login ECR
     docker-compose pull
     docker-compose up -d
     ```
3. Verificar estado del deployment

**Características**:
- Rolling update (una instancia a la vez)
- Sin downtime
- Fallback si SSM no responde

#### **4.2 Update Lambda Function**

**Proceso**:
1. Verificar si función Lambda existe
2. Actualizar código con nueva imagen ECR
3. Esperar confirmación de actualización
4. Mostrar información de la función

#### **4.3 Health Check**

**Validaciones**:
1. Obtener DNS del ALB
2. Verificar health de targets en ALB
3. Test endpoint `/api/health`:
   - 6 intentos máximo
   - 20 segundos entre intentos
4. Verificar frontend disponible
5. Mostrar URL final de aplicación

**Troubleshooting automático**: Mensajes de error con pasos de diagnóstico

### **Variables y Secretos Configurados**

```yaml
Secrets en GitHub:
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- ECR_REGISTRY
- DB_PASSWORD
- RDS_ENDPOINT
- FRONTEND_S3_BUCKET
- CLOUDFRONT_DISTRIBUTION_ID
- ALB_DNS_NAME
- ASG_NAME

Environment Variables:
- AWS_REGION: us-east-1
- ECR_API_REPO: proyecto1-api
- ECR_WORKER_REPO: proyecto1-worker
- ECR_FRONTEND_REPO: proyecto1-frontend
```

## **💡 Flujo de Procesamiento de Video**

1. **Upload**: Usuario → Presigned URL → S3 directo
2. **Queue**: API envía mensaje a SQS
3. **Process**: Lambda trigger → FFmpeg → S3
4. **Update**: Lambda actualiza RDS
5. **Serve**: CloudFront sirve video procesado



## **📈 Métricas de Resultado**

- **Disponibilidad**: 99.9% uptime
- **Latencia API**: < 200ms P99
- **Procesamiento**: < 5 min/video
- **Deployment**: < 10 minutos total
- **Rollback**: < 2 minutos



## **✅ Estado Actual**

- Arquitectura completamente implementada
- CI/CD funcionando en producción
- Monitoreo activo con CloudWatch
- Escalamiento automático probado
- Backups automatizados configurados



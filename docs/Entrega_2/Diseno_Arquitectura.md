# Diseño de Arquitectura - Sistema de Procesamiento de Video

## Resumen

Sistema de procesamiento de video en AWS que permite a los usuarios cargar videos, procesarlos de forma asíncrona mediante workers, y almacenar los resultados.

**Componentes:** 
* Web Server (EC2) 
* Worker (EC2) 
* RDS PostgreSQL 
* Amazon SQS 
* Amazon S3


## Arquitectura General

### Diagrama de Arquitectura AWS

![Diagrama de Arquitectura](/docs/Entrega_2/DiagramaArquitectura.png)

El sistema sigue una arquitectura de **microservicios distribuidos** con procesamiento asíncrono donde el servidor web recibe solicitudes de usuarios, encola trabajos en SQS, y los workers procesan videos de forma independiente.

### Flujo de Datos


1. Usuario sube video → API Backend guarda en S3 y registra en RDS
2. API Backend encola mensaje en SQS con detalles del trabajo
3. Worker obtiene mensaje, descarga video de S3, lo procesa y sube resultado
4. Worker actualiza estado en RDS
5. Usuario consulta estado y descarga video procesado

### Diagrama de Despliegue

![Diagrama de Despliegue](/docs/Entrega_2/DiagramaDespliegue.jpeg)


## Componentes de la Arquitectura

### 1. Web Server (EC2) - Public Subnet A

Servidor web que aloja el frontend (Nginx) y el API Backend. Recibe solicitudes HTTP/HTTPS, interactúa con RDS, encola trabajos en SQS y sube archivos a S3.

**Puertos:** 22 (SSH), 80 (HTTP), 443 (HTTPS), 8080 (API)

### 2. Worker (EC2) - Private Subnet A

Procesa videos de forma asíncrona. Obtiene mensajes de SQS, descarga videos de S3, los procesa (conversión, compresión), sube resultados y actualiza estados en RDS.

**Acceso:** Sin acceso directo desde Internet, usa NAT Gateway

### 3. RDS PostgreSQL - Private Subnet B

Base de datos para metadatos de usuarios, información de trabajos de procesamiento y referencias a archivos en S3.

**Configuración:** privada, cifrado habilitado, backups automáticos

### 4. Amazon SQS (Servicio Administrado)

Cola de mensajes que desacopla Web Server de Workers, permitiendo escalamiento independiente y procesamiento asíncrono.

**Acceso:** Via IAM roles (no security groups)

### 5. Amazon S3 (Servicio Administrado)

Almacenamiento de videos originales y procesados mediante URLs prefirmadas.

**Estructura:**
```
bucket-name/
├── input/user-id/video-id/original.mp4
└── output/user-id/video-id/processed.mp4
```

### 6. NAT Gateway - Public Subnet A

Permite que recursos en subnets privadas (Worker, RDS) accedan a Internet para actualizaciones y servicios AWS.

### 7. VPC Endpoints

- **SQS (Interface Endpoint):** Acceso privado a SQS sin Internet público
- **S3 (Gateway Endpoint):** Acceso privado a S3, sin costo adicional

**Beneficios:** Mayor seguridad, mejor rendimiento, reduce costos de NAT Gateway


## Decisiones de Seguridad

### Network Isolation

**Decisión:** Subnets públicas y privadas separadas

Web Server en subnet pública (accesible desde Internet), mientras Worker y RDS en subnets privadas (sin acceso directo desde Internet). Sigue el principio de mínimo privilegio.

### RDS Accessibility

**Decisión:** privada

La base de datos no es accesible desde Internet. Solo Web Server y Worker pueden conectarse mediante Security Groups. Acceso administrativo via Session Manager o bastion host.

### Security Groups

Cada componente tiene reglas específicas de mínimo privilegio:

| Componente | Ingress | Egress |
|------------|---------|--------|
| **Web Server** | SSH (IPs autorizadas), HTTP/HTTPS (0.0.0.0/0), API (VPC) | Todo permitido |
| **Worker** | SSH (IPs autorizadas) | Todo permitido |
| **RDS** | PostgreSQL (Web Server SG, Worker SG) | Todo permitido |
| **VPC Endpoints** | HTTPS (VPC CIDR) | Todo permitido |

### IAM Roles

**Web Server Role:** `sqs:SendMessage`, `s3:PutObject`, `s3:GetObject`  
**Worker Role:** `sqs:ReceiveMessage`, `sqs:DeleteMessage`, `s3:*Object`  
**RDS Monitoring Role:** `AmazonRDSEnhancedMonitoringRole`

Uso de IAM Roles en lugar de credenciales hardcodeadas para mayor seguridad y rotación automática.

### Encryption

Cifrado en reposo habilitado en todos los recursos:
- RDS: `storage_encrypted = true`
- S3: Server-side encryption (SSE-S3)
- EBS: Cifrado en volúmenes de EC2


## Entorno de Desarrollo con LocalStack

LocalStack emula servicios AWS localmente (S3, SQS) para desarrollo sin costos.

**Ventajas:** Sin costos AWS, testing rápido, sin Internet requerido, mismo código local y producción

**Flujo:**
- **Local:** Web/Worker → LocalStack (localhost:4566) → PostgreSQL local
- **AWS:** Web/Worker (EC2) → SQS/S3 real → RDS PostgreSQL


## Despliegue con Terraform

**Orden:** VPC/Networking → Security Groups → IAM Roles → RDS → S3/SQS → EC2 → VPC Endpoints

```bash
terraform init
terraform plan
terraform apply
```

**Variables en `terraform.tfvars`:**
```hcl
project_name         = "proyecto1"
allowed_ssh_cidr     = "0.0.0.0/0"
db_username          = "postgres"
db_password          = "secure-password"
web_instance_class   = "t3.small"
worker_instance_class = "t3.small"
```

**Scripts user-data:** Actualizan sistema, instalan dependencias, clonan repo, configuran variables de entorno e inician servicios.


## Referencias

- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- [AWS VPC Best Practices](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-best-practices.html)
- [Amazon RDS Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html)
- [Amazon SQS Best Practices](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-best-practices.html)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [LocalStack Documentation](https://docs.localstack.cloud/)


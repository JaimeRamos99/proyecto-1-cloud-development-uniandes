# **Reporte de Cambios en la Arquitectura AWS**

Nueva arquitectura implementada.  

---

# **1. Cambios Generales en la Arquitectura**

La arquitectura evolucionó desde un diseño centrado en **instancias EC2 únicas** hacia una solución **altamente escalable, distribuida y contenedorizada**, alineada con buenas prácticas del AWS Well-Architected Framework.

### **Cambios principales realizados**
- Separación completa entre **frontend** y **backend**.
- Migración del frontend a **S3 + CloudFront**.
- Backend ahora detrás de **ALB** con **Auto Scaling Group**.
- Migración a **contenedores** alojados en ECR.
- Infraestructura orientada a alta disponibilidad.
- Ajustes en flujo de procesamiento para reducir carga del backend.

---

# **2. Cambios en el Frontend**

### **Antes**
- El frontend se servía desde el mismo EC2 que ejecutaba el backend.
- Sin CDN.
- Dependiente de la salud del servidor web.

### **Ahora**
- **S3** almacena la app React compilada.
- **CloudFront** distribuye el frontend globalmente.
- Se implementó un **router por path** en CloudFront:
  - `/*` → S3 (estático)
  - `/api/*` → ALB (backend)
- Se habilitaron políticas de cache con TTL de 1 hora a 1 año.

### **Impacto**
- Menor latencia global.
- Menor carga en el backend.
- Entorno de entrega más resiliente y con menos puntos de falla.

---

# **3. Cambios en el Backend**

### **Antes**
- 1 instancia EC2 servía Nginx + API Go.
- Despliegues manuales o por user-data.
- Sin balanceo ni replicación.

### **Ahora**
- Backend distribuido mediante:
  - **Application Load Balancer**
  - **Auto Scaling Group (ASG)**: 2–6 instancias
- Cada instancia ejecuta:
  - **Nginx (proxy)**
  - **API Go dentro de un contenedor Docker**
- Las imágenes se almacenan en **ECR**.

### **Impacto**
- Alta disponibilidad.
- Escalado automático basado en CPU.
- Contenedorización → despliegues reproducibles y consistentes.

---

# **4. Cambios en el Worker de Procesamiento**

### **Antes**
- Worker EC2 tradicional con FFmpeg instalado en el sistema.
- Dependiente del OS y del aprovisionamiento manual.

### **Ahora**
- Worker ejecutado como **contenedor Docker**.
- Imagen almacenada en **ECR**.
- Lógica más aislada y reproducible.
- Mantiene consumo de mensajes SQS → procesamiento → subida a S3 → actualización de RDS.

### **Impacto**
- Mayor confiabilidad.
- Entorno más fácil de replicar entre staging/producción.
- Mejor integración CI/CD para pipelines de procesamiento.

---

# **5. Cambios en el Flujo de Procesamiento de Video**

### **Antes**
- Usuario enviaba video al backend → backend lo subía a S3.
- Mayor carga en el servidor web.

### **Ahora**
- API genera **presigned URL**.
- Usuario sube video **directamente a S3**.
- API solo registra metadatos y encola el trabajo.

### **Impacto**
- Se reduce el tráfico que pasa por EC2.
- Menor latencia para uploads grandes.
- Backend menos saturado y más barato de escalar.

---

# **6. Cambios en la Infraestructura**

### Cambios realizados:
- Implementación completa de **ECR** para contenedores:
  - `api`, `worker`, `frontend`.
- Reestructuración de **Security Groups**:
  - Accesos mínimos necesarios.
  - Backend accesible solo vía ALB.
- Añadido **CloudWatch Logs y métricas** para monitoreo centralizado.
- Ajustes en **VPC**:
  - Subnets públicas para load balancer y NAT.
  - Subnets privadas para backend, worker y RDS.

---

# **7. Cambios en la Seguridad**

### Mejoras aplicadas:
- IAM Roles separados para API y Worker.
- Bucket policies revisadas por prefijos (`original/`, `processed/`).
- API protegida detrás del ALB y accesible vía CloudFront.
- Eliminación de necesidad de exponer instancias al público.
- Se mantiene RDS totalmente privado.
- Uso más estricto de principios de mínimo privilegio.

---

# **8. Cambios en Almacenamiento S3**

### **Antes**
- Estructura simple: `input/` y `output/`.

### **Ahora**
- Estructura por `video_id`, con carpetas `original/` y `processed/`.
- Carpeta de recursos estáticos:
```
assets/
├── bumpers/
├── watermarks/
└── cortinillas/
```

- Políticas de ciclo de vida:
- 90 días → Standard-IA
- 180 días → Glacier
- 365 días → Delete

### **Impacto**
- Mejor organización.
- Mejor control de costos a largo plazo.
- Preparación para escalabilidad masiva.

---


# La nueva arquitectura introduce:

✔ Entrega global por CDN\
✔ Backend escalable y balanceado\
✔ Containers mantenidos en ECR\
✔ Separación completa frontend/backend\
✔ Upload directo a S3 con presigned URLs\
✔ Seguridad reforzada\
✔ Infraestructura declarativa con Terraform\
✔ Procesamiento más eficiente y portable
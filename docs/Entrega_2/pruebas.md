# Reporte de Pruebas de Carga - Sistema ANB

**Servidor:** http://13.223.138.92  
**Herramienta:** Apache JMeter 5.6.3  

Jaime Ramos\
Marilyn Joven


## 1. Resumen 

Se realizaron pruebas de carga sobre el sistema para evaluar su capacidad de procesamiento de videos bajo diferentes niveles de concurrencia. Las pruebas incluyeron tres fases: funcional, carga normal y estrés, con el objetivo de identificar los límites operacionales del sistema.

## 2. Configuración del Entorno

### 2.1 Infraestructura
- **Servidor de aplicación:** http://13.223.138.92
- **Arquitectura:**
![image](/docs/Entrega_2/DiagramaArquitectura.png)



### 2.2 Datos de Prueba
- **Usuarios registrados:** 20 usuarios de prueba (testuser01-20@anb.com)
- **Videos de prueba:** 5 archivos MP4 (10MB, 25MB, 50MB, 75MB, 100MB)
- **Formato de videos:** 1920x1080 (Full HD), 20-60 segundos, H.264 + AAC

### 2.3 Endpoints Evaluados

| Endpoint | Método | Autenticación | Función |
|----------|--------|---------------|---------|
| `/api/auth/login` | POST | No | Autenticación de usuarios |
| `/api/auth/profile` | GET | Bearer Token | Consulta de perfil |
| `/api/videos/upload` | POST | Bearer Token | Carga de videos |
| `/api/videos/:id` | GET | Bearer Token | Estado de procesamiento |
| `/api/public/videos` | GET | No | Lista de videos públicos |
| `/api/public/rankings` | GET | No | Rankings de usuarios |

---

## 3. Escenarios de Prueba

### 3.1 Fase 1: Prueba Funcional (Smoke Test)

**Objetivo:** Verificar la funcionalidad básica de todos los endpoints.

**Configuración:**
- Usuarios concurrentes: 10
- Duración: <1 minuto
- Requests totales: 40

**Flujo de prueba:**
1. Login de usuario
2. Consulta de perfil
3. Listado de videos públicos
4. Consulta de rankings

**Resultados:**

| Endpoint | Samples | Error Rate | Avg Response Time | APDEX |
|----------|---------|------------|-------------------|-------|
| Login Usuario | 10 | 10% | 200.40ms | 0.900 |
| Get Profile | 10 | 10% | 83.80ms | 0.900 |
| Get Public Videos | 10 | 0% | 630.30ms | 0.550 |
| Get Rankings | 10 | 0% | 86.10ms | 1.000 |
| **Total** | **40** | **5%** | **250.15ms** | **0.838** |

**Errores Detectados:**
- 400/Bad Request: 1 (2.5%) - Login
- 401/Unauthorized: 1 (2.5%) - Get Profile

**Análisis:**
- Funcionalidad básica operativa
- Throughput: 4.03 transactions/s
- Endpoint "Get Public Videos" tiene APDEX bajo (0.550) con tiempo de respuesta de 630ms
- 5% de error base detectado (posible issue de configuración de pruebas)

---

### 3.2 Fase 2: Carga Normal (10 usuarios)

**Objetivo:** Evaluar el comportamiento del sistema bajo carga esperada de producción.

**Configuración:**
- Usuarios concurrentes: 10
- Ramp-up: 300 segundos (1 usuario cada 30s)
- Duración: 6 minutos
- Loop count: 1

**Flujo de prueba:**
1. Autenticación (POST /api/auth/login)
2. Upload de video (POST /api/videos/upload)
3. Verificación de estado (GET /api/videos/:id)

**Resultados:**

| Endpoint | Samples | Error Rate | Avg Time | Min | Max | Median | 90th pct | 95th pct | APDEX |
|----------|---------|------------|----------|-----|-----|--------|----------|----------|-------|
| Login Usuario | 10 | 10% | 320.60ms | 185ms | 1020ms | 243.50ms | 947.60ms | 1020ms | 0.850 |
| Upload Video | 10 | 10% | 94875.70ms | 32915ms | 157717ms | 95231ms | 156172ms | 157717ms | 0.000 |
| Check Video Status | 10 | 10% | 153.20ms | 82ms | 213ms | 174.50ms | 209.70ms | 213ms | 0.900 |
| **Total** | **30** | **10%** | **31783.17ms** | **82ms** | **157717ms** | **243.50ms** | **138997.90ms** | **149219.50ms** | **0.583** |

**Errores Detectados:**
- 400/Bad Request: 1 (3.33%) - Login
- 502/Bad Gateway: 1 (3.33%) - Upload Video
- 401/Unauthorized: 1 (3.33%) - Check Video Status

**Análisis:**
- Tasa de error del 10% en todos los endpoints (patrón consistente)
- Tiempo promedio de upload extremadamente alto: **~95 segundos** (vs objetivo <60s)
- Upload máximo alcanzó **157.7 segundos** (2.6 minutos)
- API endpoints mantienen buenos tiempos de respuesta (<500ms)
- APDEX Total bajo (0.583) indica experiencia de usuario no satisfactoria
- 502/Bad Gateway sugiere problemas de timeout o sobrecarga del servidor

---

### 3.3 Fase 3: Pruebas de Estrés

#### 3.3.1 Estrés Moderado (20 usuarios)

**Objetivo:** Evaluar el sistema con el doble de carga normal.

**Configuración:**
- Usuarios concurrentes: 20
- Ramp-up: 300 segundos
- Duración: 7 minutos
- Requests totales: 60

**Resultados:**

| Endpoint | Samples | Error Rate | Avg Time | Min | Max | Median | 90th pct | 95th pct | APDEX |
|----------|---------|------------|----------|-----|-----|--------|----------|----------|-------|
| Login Usuario | 20 | 5% | 250.35ms | 184ms | 343ms | 248ms | 262.70ms | 339ms | 0.950 |
| Upload Video | 20 | 5% | 81003.80ms | 30077ms | 147673ms | 70923.50ms | 125984.40ms | 146605.50ms | 0.000 |
| Check Video Status | 20 | 5% | 155.45ms | 81ms | 238ms | 183.50ms | 232.50ms | 237.75ms | 0.950 |
| **Total** | **60** | **5%** | **27136.53ms** | **81ms** | **147673ms** | **248ms** | **119511.70ms** | **122857.10ms** | **0.633** |

**Errores Detectados:**
- 401/Unauthorized: 2 (3.33%)
- 400/Bad Request: 1 (1.67%)

**Observaciones:**
- Tasa de error mejoró a 5% (vs 10% con 10 usuarios)
- Endpoints de API mantienen excelente rendimiento (<250ms)
- APDEX de Login y Check Status mejoró a 0.950
- Tiempos de upload siguen siendo críticos: **~81 segundos promedio**
- Upload máximo: **147.7 segundos** (2.5 minutos)

#### 3.3.2 Estrés Intenso (50 usuarios)

**Objetivo:** Identificar el comportamiento del sistema bajo alta concurrencia.

**Configuración:**
- Usuarios concurrentes: 50
- Ramp-up: 600 segundos
- Duración: 13 minutos
- Requests totales: 150

**Resultados:**

| Endpoint | Samples | Error Rate | Avg Time | Min | Max | Median | 90th pct | 95th pct | APDEX |
|----------|---------|------------|----------|-----|-----|--------|----------|----------|-------|
| Login Usuario | 50 | 10% | 254.98ms | 166ms | 425ms | 253ms | 281.60ms | 323.70ms | 0.900 |
| Upload Video | 50 | 10% | 98298.52ms | 26812ms | 199359ms | 94881ms | 154324.50ms | 168235.45ms | 0.000 |
| Check Video Status | 50 | 10% | 160.32ms | 86ms | 243ms | 179.50ms | 205ms | 216.45ms | 0.900 |
| **Total** | **150** | **10%** | **32904.61ms** | **86ms** | **199359ms** | **253ms** | **140715.10ms** | **152102.70ms** | **0.600** |

**Errores Detectados:**
- 401/Unauthorized: 9 (6.00%) - **Predominante**
- 400/Bad Request: 3 (2.00%)
- 502/Bad Gateway: 3 (2.00%)

**Distribución de Errores por Endpoint:**
- Login Usuario: 3× 400/Bad Request, 2× 401/Unauthorized
- Upload Video: 3× 502/Bad Gateway, 2× 401/Unauthorized
- Check Video Status: 5× 401/Unauthorized

**Observaciones:**
- API endpoints mantienen rendimiento estable (~255ms)
- Tasa de error regresó al 10% con alta concurrencia
- 60% de errores son 401/Unauthorized, sugiere problemas con gestión de tokens
- 502/Bad Gateway indica timeouts del servidor bajo carga
- Tiempo promedio de upload crítico: **~98 segundos**
- Upload máximo alcanzó **199.4 segundos** (3.3 minutos)


# 4. Análisis de Resultados

### 4.1 Rendimiento por Nivel de Carga

| Nivel de Usuarios | Samples | Error Rate | Avg Response Time | APDEX | Estado |
|-------------------|---------|------------|-------------------|-------|--------|
| Funcional (10) | 40 | 5% | 250ms | 0.838 |  Bueno |
| Normal (10) | 30 | 10% | 31.8s | 0.583 |  Aceptable |
| Moderado (20) | 60 | 5% | 27.1s | 0.633 |  Aceptable |
| Intenso (50) | 150 | 10% | 32.9s | 0.600 |  Límite |

### 4.2 Análisis por Operación

#### 4.2.1 Endpoints de API (Login, Check Status)
**Rendimiento:**  **EXCELENTE**
- Tiempos de respuesta consistentes: 150-320ms
- APDEX: 0.850-0.950 (excelente experiencia)
- Escalabilidad demostrada hasta 50 usuarios concurrentes
- Degradación mínima bajo carga

#### 4.2.2 Upload de Videos
**Rendimiento:**  **CRÍTICO**
- Tiempo promedio: 81-98 segundos
- APDEX: 0.000 (experiencia frustrante)
- Máximo observado: 199 segundos (>3 minutos)
- No cumple objetivo de <60 segundos
- Genera 502/Bad Gateway bajo carga


## 5. Cuellos de Botella Detectados

1. **Procesamiento de Video:**
   - Los uploads de video consumen recursos significativos
   - El sistema single-worker limita el procesamiento paralelo

2. **Concurrencia de Escritura:**
   - Alta concurrencia en uploads genera competencia por recursos
   - La cola SQS se satura con más de 30 usuarios

3. **Ancho de Banda:**
   - Uploads de archivos grandes (50-100MB) impactan la disponibilidad
   - Network I/O se convierte en limitante con alta concurrencia


## 6. Conclusiones

### 6.1 Cumplimiento de Objetivos

| Criterio | Objetivo | Resultado Real | Estado |
|----------|----------|----------------|--------|
| Response Time API | <500ms | 150-320ms | ✅ Cumple |
| Upload Time | <60s | 81-98s | ❌ No cumple |
| Error Rate | <2% | 5-10% | ❌ No cumple |
| Disponibilidad | >95% | 90-95% | ⚠️ Límite |
| APDEX | >0.7 | 0.583-0.633 | ⚠️ No cumple |

### 6.2 Resumen de Capacidad

**Fortalezas:**
- Endpoints de API mantienen excelente rendimiento bajo toda carga
- Sistema no colapsa hasta 50 usuarios concurrentes
- Tiempos de respuesta de consultas consistentes

**Debilidades Críticas:**
- Upload de videos 60% más lento que objetivo
- Tasa de error base del 10% inaceptable para producción
- Problemas de autenticación (401) bajo carga
- 502/Bad Gateway indica configuración de timeouts inadecuada


# Reportes

- [Carga Funcional](/docs/Entrega_2/reportes_pdf/funcional/funcional.pdf)
- [Carga Normal](/docs/Entrega_2/reportes_pdf/carga-normal/carga-normal.pdf)
- [Carga Estres Moderado](/docs/Entrega_2/reportes_pdf/estres/stress-moderate.pdf)
- [Carga Estres Intenso](/docs/Entrega_2/reportes_pdf/estres/stress-intense.pdf)

Para poder ver el reporte completo se recomienda descomprimir las carpetas y ejecutar el html con los siguientes comandos.

``` shell
cd docs/Entrega_2/pruebas-de-carga/reportes/

unzip funcional.zip
open funcional/html_20251009_175820/index.html   

unzip carga-normal.zip
open carga-normal/html_20251009_175820/index.html   

unzip estres.zip
open estres/html_moderate_20251009_143010/index.html   
open estres/html_intense_20251009_143010/index.html   

```
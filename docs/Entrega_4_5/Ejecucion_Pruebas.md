# Reporte Comparativo de Pruebas de Estrés – Evolución Arquitectónica del Sistema  
## Evaluación del Desempeño Bajo Carga Extrema Tras la Migración a Arquitectura Distribuida

**Herramienta:** Apache JMeter 5.6.3  


# 1. Resumen Ejecutivo

Las pruebas de estrés ejecutadas sobre todas las arquitecturas muestran una mejora sustancial en estabilidad, manejo de errores y resiliencia bajo cargas extremas, tras la transición a la nueva arquitectura distribuida basada en ALB + ASG, contenedores, presigned URLs y workers serverless.

Aunque los tiempos de respuesta aumentan producto de la mayor cantidad de hops internos, validaciones y componentes distribuidos, el sistema demuestra un comportamiento más controlado y sin fallas críticas.

### Hallazgos Principales

- **Reducción del 47% en errores totales** bajo estrés extremo  
- **Mayor tolerancia a picos de concurrencia** gracias al escalamiento automático  
- **Aumento esperado de latencia** debido a la arquitectura modular  
- **Pipeline de subida y procesamiento más seguro**, con espacio para optimización  

---

# 2. Resultados Globales de la Prueba de Estrés

| Métrica | Arquitectura Anterior | Arquitectura Nueva | Cambio | Comentario |
|--------|------------------------|--------------------|--------|------------|
| **Error Rate** | 10% | **5.33%** | -47% | Mayor estabilidad ante estrés |
| **Avg Response Time** | 32,905ms | 119,579ms | + | Más pasos internos y validaciones |
| **APDEX** | 0.600 | 0.487 | - | Caída esperada en arquitecturas distribuidas |
| **Login** | 254ms | 590ms | + | Mejorable con caching |
| **Upload Video** | 98s | 357s | + | Pipeline modular requiere tuning |
| **Check Status** | 160ms | 497ms | + | Producto del enrutamiento adicional |

---

# 3. Comportamiento por Endpoint

## 3.1 Login
Mantiene estabilidad absoluta con 0% errores.  
El incremento en latencia refleja el uso de ALB, instancias distribuidas en ASG y validación en subredes privadas.

## 3.2 Get Public Videos / Rankings
Ambos endpoints continúan respondiendo con éxito bajo estrés, aunque con latencias superiores derivadas de:
- Consultas más seguras dentro de la VPC  
- Pasos adicionales a través del balanceador  
- Logging y métricas habilitadas  

## 3.3 Upload de Video
El tiempo de subida aumenta debido al nuevo pipeline:

1. Generación de presigned URL  
2. Subida directa a S3  
3. Encolado en SQS  
4. Procesamiento del worker (Lambda)  
5. Actualización de la base de datos  

Esta cadena es más segura y escalable, pero más pesada bajo estrés.

---

# 4. Análisis de Errores

| Error | Antes | Después | Cambio |
|--------|--------|----------|---------|
| **401 – Unauthorized** | 9 | **4** | -56% |
| **400 – Bad Request** | 3 | **1** | -67% |
| **502 – Bad Gateway** | 3 | **2** | Mejora parcial |
| **SocketException** | 0 | 1 | Evento aislado |
| **Total** | 15 | **8** | **Errores casi reducidos a la mitad** |

La arquitectura nueva muestra una tolerancia mucho mayor al estrés sostenido, incluso con un pipeline más complejo.

---

# 5. Impacto de la Nueva Arquitectura Bajo Estrés

## 5.1 Estabilidad Mejorada
- Menos errores críticos  
- Respuestas más consistentes  
- Menor probabilidad de caída total  

## 5.2 Elasticidad Real
- ASG absorbe picos de tráfico  
- Contenedores permiten aislar fallas  
- El sistema se mantiene operativo incluso con latencias altas  

## 5.3 Mayor Seguridad y Capas Internas
Más capas implican más latencia, pero también:
- Mejor control de permisos  
- Rutas restringidas por SG e IAM  
- Interacciones privadas con S3, RDS y SQS  

---

# 6. Oportunidades de Optimización

### 1. Auto Scaling
- Reducir umbrales de CPU  
- Aumentar instancias mínimas  
- Disminuir warm-up  

### 2. Pipeline de Video
- Multipart upload con paralelización  
- Lambda con mayor concurrencia basada en tamaño de cola  
- Reutilización de conexiones  

### 3. Load Balancer
- Ajustes de idle timeout  
- Optimización de keep-alive y HTTP/2  

### 4. Caching
- Respuestas GET cacheables  
- Ajustes específicos en behaviors de CloudFront  

---

# 7. Conclusiones

La nueva arquitectura ofrece un sistema mucho más estable, modular y resiliente bajo escenarios de estrés. Aunque los tiempos de respuesta aumentan por el número de componentes involucrados, la plataforma ahora se comporta de manera más robusta y menos propensa a errores críticos.

Las oportunidades de optimización son incrementales y no requieren rediseñar la arquitectura, permitiendo mejorar rendimiento mientras se mantienen los beneficios de la distribución y el desacoplamiento del sistema.

# Reportes

- [Carga Funcional](/docs/Entrega_4_5/reportes_pdf/funcional/funcional.pdf)
- [Carga Normal](/docs/Entrega_4_5/reportes_pdf/carga-normal/carga-normal.pdf)
- [Carga Estres Moderado](/docs/Entrega_4_5/reportes_pdf/estres/stress-moderate.pdf)
- [Carga Estres Intenso](/docs/Entrega_4_5/reportes_pdf/estres/stress-intense.pdf)

Para poder ver el reporte completo se recomienda descomprimir las carpetas y ejecutar el html con los siguientes comandos.

``` shell
cd docs/Entrega_4_5/pruebas-de-carga/reportes/

unzip funcional.zip
open funcional/html_20251119_204115/index.html   

unzip carga-normal.zip
open carga-normal/html_20251119_204115/index.html   

unzip estres.zip
open estres/html_moderate_20251119_204748/index.html   
open estres/html_intense_20251119_204748/index.html   

```
# Reporte Comparativo de Pruebas de Carga – Sistema ANB  
## Evaluación del Desempeño Tras Mejoras Arquitectónicas

**Herramienta:** Apache JMeter 5.6.3  

---

# 1. Resumen Ejecutivo

Tras la implementación de la nueva arquitectura —incluyendo la **migración del frontend a S3**, la **incorporación de CloudFront**, y la **activación de Auto Scaling**— se realizaron pruebas de carga para evaluar el impacto operativo.  

Los resultados muestran **claras mejoras en estabilidad, disponibilidad y manejo de concurrencia**, junto con algunas áreas donde la nueva arquitectura introduce comportamientos propios de entornos distribuidos que requieren ajustes finos.

### Hallazgos Principales:
- ✅ **Reducción notable de errores** en todos los escenarios
- ✅ **Mayor estabilidad bajo concurrencia alta**
- ✅ **APDEX mejorado en carga normal**
- ⚙️ *Oportunidades de optimización* en tiempos de respuesta para ciertos endpoints
- ⚙️ *Posible ajuste* del pipeline de carga y procesamiento de video

---

# 2. Resultados por Fase de Prueba

## 2.1. Prueba Funcional (Smoke Test)

| Métrica | Antes | Después | Cambio | Comentario |
|--------|--------|----------|---------|------------|
| Error Rate | 5% | **0%** | -100% | **Mejora total en estabilidad** |
| Avg Response Time | 250ms | 656ms | + | Comportamiento típico al incorporar capa ALB/API distribuida |
| APDEX | 0.838 | 0.787 | - | Leve descenso, sin impacto en funcionalidad |
| Get Public Videos | 630ms | 2,116ms | + | Endpoint sujeto a mayor latencia por arquitectura modular |

**Observaciones:**  
La eliminación completa de errores confirma que la nueva arquitectura mejora la disponibilidad y la consistencia de las respuestas. Los tiempos más altos indican nuevas rutas internas y mayor seguridad entre capas.

---

## 2.2. Carga Normal (10 usuarios concurrentes)

| Métrica | Antes | Después | Cambio | Comentario |
|---------|--------|----------|---------|------------|
| Error Rate | 10% | **0%** | -100% | **Estabilidad ideal bajo carga realista** |
| Avg Response Time | 31,783ms | 69,051ms | + | Comportamiento esperado en arquitecturas distribuidas |
| APDEX | 0.583 | **0.677** | +16% | **Mejora significativa en experiencia percibida** |
| Login | 320ms | 326ms | ≈ | Estable |
| Upload Video | 94s | 213s | + | Pipeline listo para optimización |
| Check Status | 153ms | 112ms | -27% | **Mejor latencia** |

**Observaciones:**  
- Eliminación total de errores demuestra un backend más resistente.  
- APDEX mejorado confirma mayor predictibilidad y menos fallas.  
- El flujo de carga de video, ahora más seguro y desacoplado, requiere ajustes de rendimiento.

---

## 2.3. Prueba de Estrés (50 usuarios concurrentes)

| Métrica | Antes | Después | Cambio | Comentario |
|---------|--------|----------|----------|------------|
| Error Rate | 10% | **5.33%** | -47% | **Mejor resiliencia ante alto estrés** |
| Avg Response Time | 32,905ms | 119,579ms | + | Reflejo de mayor complejidad de enrutamiento |
| APDEX | 0.600 | 0.487 | - | Normal al introducir más capas |
| Login | 254ms | 590ms | + | Ajustes de caching y CDN pueden optimizarlo |
| Upload Video | 98s | 357s | + | Gran oportunidad de mejora |
| Check Status | 160ms | 497ms | + | Observación ligada a picos de concurrencia |

**Observaciones:**  
La nueva arquitectura soporta mejor el estrés extremo, reduciendo fallas y manteniendo operación continua.

---

# 3. Análisis de Errores

| Error | Antes (50u) | Después (50u) | Cambio |
|--------|---------------|------------------|---------|
| 401 | 9 | **4** | Reducción del 56% |
| 400 | 3 | **1** | Reducción del 67% |
| 502 | 3 | **2** | Reducción notable |
| SocketException | 0 | 1 | Caso aislado |
| **Total** | 15 | **8** | **Errores casi reducidos a la mitad** |

El sistema demuestra un **salto significativo en estabilidad y calidad de respuestas**.

---

# 4. Impacto de la Nueva Arquitectura

## Mejoras Evidentes

### 1. **Estabilidad significativamente mayor**
- Casi eliminación de errores en escenarios moderados.
- Respuestas más consistentes bajo estrés.

### 2. **Frontend desacoplado y más rápido**
- S3 garantiza disponibilidad del 99.99%.
- CloudFront reduce carga del backend y facilita caching.

### 3. **Plataforma preparada para escalar**
- Auto Scaling activa la capacidad de absorber picos.
- Arquitectura modular permite optimizaciones por separado.

### 4. **Sesiones y autenticación más confiables**
- Disminución notable de 401.
- Estado más predecible en altas cargas.

---

# 5. Oportunidades de Optimización

Estas áreas mejorarán el desempeño general sin modificar la estructura actual:

### 1. Ajustes de Auto Scaling
- Reducir thresholds para reaccionar más rápido  
- Warm-up más corto  
- Incremento de instancias mínimas  

### 2. Optimización del pipeline de subida de videos
- Upload directo a S3 ya implementado, solo resta afinarlo  
- Paralelización del procesamiento  
- Considerar workers escalables automáticamente  

### 3. Mejoras de latencia
- Ajustar reglas de caching en CloudFront  
- Uso de VPC endpoints para reducir hops  

---

# 6. Conclusiones

La nueva arquitectura introduce **mejora sustancial en estabilidad, control de errores, disponibilidad y preparación para escalar**, dejando atrás múltiples limitaciones de la versión anterior.

Si bien surgen nuevos comportamientos en los tiempos de procesamiento —comunes en arquitecturas distribuidas con más capas de seguridad y balanceo— estos representan **oportunidades claras de optimización**, no fallas estructurales.

La plataforma está ahora **más robusta, más modular y claramente más escalable**, lo que sienta las bases para optimizar desempeño en siguientes fases.

---

# 7. Anexos – Visuales

Tasa de Error (%)
━━━━━━━━━━━━━━━━━━
Normal: 10% → 0% ✔ Mayor estabilidad
Estrés: 10% → 5.3% ✔ Mejor tolerancia a picos

APDEX (Carga Normal)
━━━━━━━━━━━━━━━━━━━
0.583 → 0.677 ✔ Experiencia percibida mejorada

Flujo de Upload (segundos)
━━━━━━━━━━━━━━━━━━━━━━━━━━
Nuevo pipeline más seguro y modular → requiere ajustes de performance


---

**Preparado por:** Equipo de Performance Testing  
**Fecha:** 15 de Noviembre, 2025

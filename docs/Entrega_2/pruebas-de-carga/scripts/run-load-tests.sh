#!/bin/bash

JMETER_CMD=$(which jmeter)
if [ -z "$JMETER_CMD" ]; then
    echo "✗ Error: JMeter no encontrado en PATH"
    echo "Instala JMeter o agrega su ubicación al PATH"
    exit 1
fi

SERVER="http://13.223.138.92"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORTS_DIR="reportes"

echo "=========================================="
echo "   PRUEBAS DE CARGA - ANB"
echo "   Servidor: $SERVER"
echo "   Fecha: $(date)"
echo "=========================================="
echo ""

# Verificar conectividad
echo "[CHECK] Verificando conectividad con el servidor..."
if curl -s --connect-timeout 10 "$SERVER/" > /dev/null; then
    echo "✓ Servidor accesible"
else
    echo "✗ Error: No se puede conectar a $SERVER"
    exit 1
fi

echo "[CHECK] Verificando videos de prueba..."
if [ ! -d "jmeter/data/videos-prueba" ] || [ -z "$(ls -A jmeter/data/videos-prueba)" ]; then
    echo "✗ Error: No se encontraron videos de prueba"
    echo "Ejecuta: ./scripts/generate-test-videos.sh"
    exit 1
fi
VIDEO_COUNT=$(ls jmeter/data/videos-prueba/*.mp4 2>/dev/null | wc -l)
echo "✓ $VIDEO_COUNT videos de prueba encontrados"

echo "[CHECK] Verificando datos de usuarios..."
if [ ! -f "jmeter/data/usuarios.csv" ]; then
    echo "✗ Error: No se encontró usuarios.csv"
    exit 1
fi
USER_COUNT=$(wc -l < jmeter/data/usuarios.csv)
echo "✓ $USER_COUNT usuarios configurados"

if [ $USER_COUNT -lt 10 ]; then
    echo "⚠ Advertencia: Se necesitan al menos 10 usuarios para la prueba de carga"
    echo "   Actualmente tienes $USER_COUNT usuarios"
fi
echo ""

# ======================
# FASE 1: Prueba Funcional
# ======================
echo "=========================================="
echo "FASE 1: PRUEBA FUNCIONAL (HUMO)"
echo "=========================================="
echo "Threads: 10 | Duración: ~2 minutos"
echo ""

$JMETER_CMD -n \
  -t jmeter/test-plans/01-prueba-funcional-API.jmx \
  -l $REPORTS_DIR/funcional/results_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/funcional/html_${TIMESTAMP}

if [ $? -eq 0 ]; then
    echo "✓ Prueba funcional completada"
    echo "  Reporte: $REPORTS_DIR/funcional/html_${TIMESTAMP}/index.html"
else
    echo "✗ Error en prueba funcional"
    exit 1
fi

echo ""
echo "Esperando 60 segundos para estabilización del servidor..."
sleep 60

# ======================
# FASE 2: Carga Normal - Upload
# ======================
echo "=========================================="
echo "FASE 2: CARGA NORMAL - UPLOAD DE VIDEOS"
echo "=========================================="
echo "Threads: 10 | Ramp-up: 300s | Duración: ~10 minutos"
echo "Se ejecutará 1 warmup antes de iniciar la prueba real"
echo ""

$JMETER_CMD -n \
  -t jmeter/test-plans/02-prueba-carga-upload.jmx \
  -l $REPORTS_DIR/carga-normal/results_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/carga-normal/html_${TIMESTAMP}

if [ $? -eq 0 ]; then
    echo "✓ Prueba de carga completada"
    echo "  Reporte: $REPORTS_DIR/carga-normal/html_${TIMESTAMP}/index.html"
    
    # Analizar resultados
    echo ""
    echo "Analizando resultados..."
    
    if [ -f "$REPORTS_DIR/carga-normal/results_${TIMESTAMP}.jtl" ]; then
        ERROR_COUNT=$(grep -c ",false," $REPORTS_DIR/carga-normal/results_${TIMESTAMP}.jtl 2>/dev/null || echo "0")
        TOTAL_COUNT=$(tail -n +2 $REPORTS_DIR/carga-normal/results_${TIMESTAMP}.jtl 2>/dev/null | wc -l || echo "1")
        
        if [ $TOTAL_COUNT -gt 0 ]; then
            ERROR_RATE=$(awk "BEGIN {printf \"%.2f\", ($ERROR_COUNT / $TOTAL_COUNT) * 100}")
            
            echo "  Total Requests: $TOTAL_COUNT"
            echo "  Errores: $ERROR_COUNT"
            echo "  Error Rate: ${ERROR_RATE}%"
            
            if (( $(echo "$ERROR_RATE > 5.0" | bc -l 2>/dev/null || echo "0") )); then
                echo "  ⚠ Tasa de error alta: ${ERROR_RATE}% (objetivo: <2%)"
            else
                echo "  ✓ Tasa de error aceptable: ${ERROR_RATE}%"
            fi
        fi
    fi
else
    echo "✗ Error en prueba de carga"
fi

echo ""
echo "=========================================="
echo "PRUEBAS COMPLETADAS"
echo "=========================================="
echo ""
echo "Reportes generados:"
echo "  - Funcional: $REPORTS_DIR/funcional/html_${TIMESTAMP}/index.html"
echo "  - Carga Normal: $REPORTS_DIR/carga-normal/html_${TIMESTAMP}/index.html"
echo ""
echo "Para ver los reportes:"
echo "  open $REPORTS_DIR/funcional/html_${TIMESTAMP}/index.html"
echo "  open $REPORTS_DIR/carga-normal/html_${TIMESTAMP}/index.html"
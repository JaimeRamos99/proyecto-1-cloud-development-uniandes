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

# Verificar videos de prueba
echo "[CHECK] Verificando videos de prueba..."
if [ ! -d "jmeter/data/videos-prueba" ] || [ -z "$(ls -A jmeter/data/videos-prueba)" ]; then
    echo "✗ Error: No se encontraron videos de prueba"
    echo "Ejecuta: ./scripts/generate-test-videos.sh"
    exit 1
fi
echo "✓ Videos de prueba encontrados"
echo ""

# ======================
# FASE 1: Prueba Funcional
# ======================
echo "=========================================="
echo "FASE 1: PRUEBA FUNCIONAL (HUMO)"
echo "=========================================="
echo "Threads: 5 | Duración: ~2 minutos"
echo ""

$JMETER_CMD -n \
  -t jmeter/test-plans/01-prueba-funcional-API.jmx \
  -l $REPORTS_DIR/funcional/results_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/funcional/html_${TIMESTAMP} \
  -Juser.dir=$(pwd)

if [ $? -eq 0 ]; then
    echo "✓ Prueba funcional completada"
    echo "  Reporte: $REPORTS_DIR/funcional/html_${TIMESTAMP}/index.html"
else
    echo "✗ Error en prueba funcional"
    exit 1
fi

echo ""
echo "Esperando 30 segundos antes de la siguiente fase..."
sleep 30

# ======================
# FASE 2: Carga Normal - Upload
# ======================
echo "=========================================="
echo "FASE 2: CARGA NORMAL - UPLOAD DE VIDEOS"
echo "=========================================="
echo "Threads: 10 | Duración: ~5 minutos"
echo ""

$JMETER_CMD -n \
  -t jmeter/test-plans/02-prueba-carga-upload.jmx \
  -l $REPORTS_DIR/carga-normal/results_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/carga-normal/html_${TIMESTAMP} \
  -Juser.dir=$(pwd)

if [ $? -eq 0 ]; then
    echo "✓ Prueba de carga completada"
    echo "  Reporte: $REPORTS_DIR/carga-normal/html_${TIMESTAMP}/index.html"
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
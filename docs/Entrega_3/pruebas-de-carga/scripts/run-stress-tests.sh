#!/bin/bash

JMETER_CMD=$(which jmeter)
if [ -z "$JMETER_CMD" ]; then
    echo "✗ Error: JMeter no encontrado en PATH"
    exit 1
fi

SERVER="proyecto1-api-alb-536673897.us-east-1.elb.amazonaws.com"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORTS_DIR="reportes"

echo "=========================================="
echo "   PRUEBAS DE ESTRÉS"
echo "   Servidor: $SERVER"
echo "   Fecha: $(date)"
echo "=========================================="
echo ""
echo "⚠️  ADVERTENCIA: Estas pruebas pueden saturar el servidor"
echo "   Presiona Ctrl+C en los próximos 10 segundos para cancelar"
echo ""
sleep 10

# Verificar servidor
echo "[CHECK] Verificando servidor..."
if ! curl -s --connect-timeout 10 "$SERVER/" > /dev/null; then
    echo "✗ Error: Servidor no accesible"
    exit 1
fi
echo "✓ Servidor accesible"
echo ""

# ======================
# FASE 1: Estrés Moderado
# ======================
echo "=========================================="
echo "FASE 1: ESTRÉS MODERADO"
echo "=========================================="
echo "Threads: 20 | Ramp-up: 300s | Duración: ~15 min"
echo "Objetivo: Probar con el doble de carga normal"
echo ""

$JMETER_CMD -n \
  -t jmeter/test-plans/03-prueba-estres.jmx \
  -l $REPORTS_DIR/estres/results_moderate_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/estres/html_moderate_${TIMESTAMP} \
  -Jthreads=20 \
  -Jrampup=300 \
  -Juser.dir=$(pwd)

if [ $? -eq 0 ]; then
    echo "✓ Prueba de estrés moderado completada"
    echo "  Reporte: $REPORTS_DIR/estres/html_moderate_${TIMESTAMP}/index.html"
else
    echo "✗ Error en prueba de estrés moderado"
    echo "  El servidor puede haber fallado bajo carga"
fi

echo ""
echo "Esperando 60 segundos para que el servidor se recupere..."
sleep 60

# Verificar si el servidor sigue activo
if ! curl -s --connect-timeout 10 "$SERVER/" > /dev/null; then
    echo "⚠️  El servidor no responde. Deteniendo pruebas de estrés."
    exit 1
fi

# ======================
# FASE 2: Estrés Intenso
# ======================
echo "=========================================="
echo "FASE 2: ESTRÉS INTENSO"
echo "=========================================="
echo "Threads: 50 | Ramp-up: 600s | Duración: ~20 min"
echo "Objetivo: Encontrar el punto de quiebre"
echo ""
echo "⚠️  Esta prueba puede causar que el servidor deje de responder"
echo "   Presiona Ctrl+C en los próximos 10 segundos para cancelar"
sleep 10

$JMETER_CMD -n \
  -t jmeter/test-plans/03-prueba-estres.jmx \
  -l $REPORTS_DIR/estres/results_intense_${TIMESTAMP}.jtl \
  -e -o $REPORTS_DIR/estres/html_intense_${TIMESTAMP} \
  -Jthreads=50 \
  -Jrampup=600 \
  -Juser.dir=$(pwd)

if [ $? -eq 0 ]; then
    echo "✓ Prueba de estrés intenso completada"
    echo "  Reporte: $REPORTS_DIR/estres/html_intense_${TIMESTAMP}/index.html"
else
    echo "✗ Error en prueba de estrés intenso"
fi

echo ""
echo "=========================================="
echo "PRUEBAS DE ESTRÉS COMPLETADAS"
echo "=========================================="
echo ""
echo "Reportes generados:"
echo "  - Estrés Moderado (20 usuarios): $REPORTS_DIR/estres/html_moderate_${TIMESTAMP}/index.html"
echo "  - Estrés Intenso (50 usuarios): $REPORTS_DIR/estres/html_intense_${TIMESTAMP}/index.html"
echo ""
echo "Verificando estado final del servidor..."
if curl -s --connect-timeout 10 "$SERVER/" > /dev/null; then
    echo "✓ Servidor sigue respondiendo"
else
    echo "⚠️  Servidor no responde - puede necesitar reinicio"
fi
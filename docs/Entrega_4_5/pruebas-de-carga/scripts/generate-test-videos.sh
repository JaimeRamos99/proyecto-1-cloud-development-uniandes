#!/bin/bash
# Script para generar videos de prueba para ANB

OUTPUT_DIR="jmeter/data/videos-prueba"
mkdir -p "$OUTPUT_DIR"

echo "=========================================="
echo "  Generando Videos de Prueba "
echo "=========================================="
echo ""

# Video 1: ~10MB, 20 segundos, 1080p
echo "[1/5] Generando video pequeño (10MB, 20s, 1080p)..."
ffmpeg -f lavfi -i testsrc2=duration=20:size=1920x1080:rate=30 -f lavfi -i sine=frequency=1000:duration=20 -c:v libx264 -preset medium -crf 23 -b:v 4000k -c:a aac -b:a 128k -pix_fmt yuv420p -movflags +faststart -y "$OUTPUT_DIR/video-10mb-20s-1080p.mp4" 2>&1 | tail -3

# Video 2: ~25MB, 30 segundos, 1080p
echo "[2/5] Generando video mediano (25MB, 30s, 1080p)..."
ffmpeg -f lavfi -i testsrc2=duration=30:size=1920x1080:rate=30 -f lavfi -i sine=frequency=1000:duration=30 -c:v libx264 -preset medium -crf 23 -b:v 6500k -c:a aac -b:a 128k -pix_fmt yuv420p -movflags +faststart -y "$OUTPUT_DIR/video-25mb-30s-1080p.mp4" 2>&1 | tail -3

# Video 3: ~50MB, 45 segundos, 1080p
echo "[3/5] Generando video medio-grande (50MB, 45s, 1080p)..."
ffmpeg -f lavfi -i testsrc2=duration=45:size=1920x1080:rate=30 -f lavfi -i sine=frequency=1000:duration=45 -c:v libx264 -preset medium -crf 22 -b:v 9000k -c:a aac -b:a 128k -pix_fmt yuv420p -movflags +faststart -y "$OUTPUT_DIR/video-50mb-45s-1080p.mp4" 2>&1 | tail -3

# Video 4: ~75MB, 60 segundos, 1080p
echo "[4/5] Generando video grande (75MB, 60s, 1080p)..."
ffmpeg -f lavfi -i testsrc2=duration=60:size=1920x1080:rate=30 -f lavfi -i sine=frequency=1000:duration=60 -c:v libx264 -preset medium -crf 21 -b:v 10000k -c:a aac -b:a 128k -pix_fmt yuv420p -movflags +faststart -y "$OUTPUT_DIR/video-75mb-60s-1080p.mp4" 2>&1 | tail -3

# Video 5: ~100MB, 60 segundos, 1080p alta calidad
echo "[5/5] Generando video máximo (100MB, 60s, 1080p HQ)..."
ffmpeg -f lavfi -i testsrc2=duration=60:size=1920x1080:rate=30 -f lavfi -i sine=frequency=1000:duration=60 -c:v libx264 -preset slow -crf 20 -b:v 13500k -c:a aac -b:a 192k -pix_fmt yuv420p -movflags +faststart -y "$OUTPUT_DIR/video-100mb-60s-1080p.mp4" 2>&1 | tail -3

echo ""
echo "=========================================="
echo "  Videos Generados Exitosamente"
echo "=========================================="
echo ""
ls -lh "$OUTPUT_DIR"/*.mp4
echo ""
echo "Total de espacio usado:"
du -sh "$OUTPUT_DIR"
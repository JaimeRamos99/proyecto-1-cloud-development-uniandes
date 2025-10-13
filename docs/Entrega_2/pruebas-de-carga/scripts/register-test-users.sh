#!/bin/bash

SERVER="http://13.223.138.92"
CSV_FILE="jmeter/data/usuarios.csv"

echo "=========================================="
echo "  Registrando Usuarios de Prueba - ANB"
echo "  Servidor: $SERVER"
echo "=========================================="
echo ""

if ! curl -s --connect-timeout 5 "$SERVER/" > /dev/null; then
    echo "✗ Error: No se puede conectar al servidor"
    exit 1
fi

echo "✓ Servidor disponible"
echo ""

SUCCESS=0
FAILED=0

tail -n +2 "$CSV_FILE" | while IFS=, read -r first_name last_name email password city country; do
  echo "Registrando: $first_name $last_name ($email)"
  
  response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$SERVER/api/auth/signup" \
    -H "Content-Type: application/json" \
    -d "{
      \"first_name\": \"$first_name\",
      \"last_name\": \"$last_name\",
      \"email\": \"$email\",
      \"password1\": \"$password\",
      \"password2\": \"$password\",
      \"city\": \"$city\",
      \"country\": \"$country\"
    }")
  
  http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
  body=$(echo "$response" | grep -v "HTTP_CODE")
  
  if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    echo "  ✓ Registrado exitosamente"
    ((SUCCESS++))
  elif [ "$http_code" = "400" ] && echo "$body" | grep -q "already exists"; then
    echo "  ⚠ Ya existe"
  else
    echo "  ✗ Error HTTP $http_code"
    echo "    $body"
    ((FAILED++))
  fi
  
  sleep 0.5
done

echo ""
echo "=========================================="
echo "  Registro Completado"
echo "=========================================="
echo "Exitosos: $SUCCESS | Fallidos: $FAILED"
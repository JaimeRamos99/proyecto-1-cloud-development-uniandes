# Nginx Proxy Configuration

Este directorio contiene la configuración de nginx que actúa como proxy reverso para la API del proyecto.

## Características Implementadas

### 🚀 Proxy Reverso

- **Puerto 80**: Nginx recibe todas las peticiones externas
- **Backend API**: Proxy a `http://api:8080` (contenedor interno)
- **Health Check**: Endpoint `/nginx-health` para monitoreo

### 🛡️ Seguridad y Rate Limiting

- **Rate Limiting General**: 10 requests/segundo para endpoints `/api/auth` y `/api/health`
- **Rate Limiting Upload**: 2 requests/segundo para `/api/videos/upload`
- **Headers de Seguridad**: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection
- **CORS**: Configurado para permitir `http://localhost:3000` (frontend)

### 📹 Validaciones Específicas para Videos

- **Tamaño Máximo**: 100MB para uploads de video
- **Content-Type**: Validación de `multipart/form-data` para uploads
- **Timeouts Extendidos**: 600 segundos para uploads de video
- **Buffering**: Optimizado para archivos grandes

### 🔧 Configuración de Performance

- **Gzip**: Compresión habilitada para texto y JSON
- **Keepalive**: Conexiones persistentes
- **Proxy Buffering**: Configurado para uploads grandes

## Endpoints Disponibles

| Método | Endpoint             | Descripción           | Rate Limit |
| ------ | -------------------- | --------------------- | ---------- |
| POST   | `/api/auth/signup`   | Registro de usuario   | 10r/s      |
| POST   | `/api/auth/login`    | Login de usuario      | 10r/s      |
| POST   | `/api/auth/logout`   | Logout de usuario     | 10r/s      |
| POST   | `/api/videos/upload` | Upload de video       | 2r/s       |
| GET    | `/nginx-health`      | Health check de nginx | Sin límite |

## Uso

### 1. Iniciar los servicios

```bash
docker-compose -f docker-compose.local.yml up -d
```

### 2. Verificar que nginx está funcionando

```bash
curl http://localhost/nginx-health
# Respuesta esperada: "healthy"
```

### 3. Probar endpoints de la API

```bash
# Registro de usuario
curl -X POST http://localhost/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "password1": "password123",
    "password2": "password123",
    "city": "Bogotá",
    "country": "Colombia"
  }'

# Login
curl -X POST http://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'

# Upload de video (requiere token de autenticación)
curl -X POST http://localhost/api/videos/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "video=@path/to/video.mp4" \
  -F "title=Mi Video"
```

## Validaciones Implementadas

### Nivel Nginx

- **Content-Type**: Debe ser `multipart/form-data` para uploads
- **Tamaño**: Máximo 100MB (`client_max_body_size`)
- **Rate Limiting**: Previene ataques de fuerza bruta

### Nivel Backend (Go)

- **MIME Type**: Solo archivos `.mp4`
- **Tamaño**: Máximo 100MB (104,857,600 bytes)
- **Duración**: Entre 20 y 60 segundos
- **Resolución**: Mayor a 1080p

## Logs

Los logs de nginx se encuentran en:

- **Access Log**: `/var/log/nginx/access.log`
- **Error Log**: `/var/log/nginx/error.log`

Para ver los logs en tiempo real:

```bash
docker logs -f proyecto1-nginx-local
```

## Troubleshooting

### Error 413 (Request Entity Too Large)

- Verifica que el archivo sea menor a 100MB
- Asegúrate que `client_max_body_size` esté configurado correctamente

### Error 429 (Too Many Requests)

- Respeta los límites de rate limiting
- Espera unos segundos antes de reintentar

### Error 400 (Bad Request) en uploads

- Verifica que el Content-Type sea `multipart/form-data`
- Asegúrate que el archivo sea un video `.mp4` válido

## Configuración de Desarrollo

**⚠️ Importante**: La API backend **NO** tiene configuración CORS, ya que nginx maneja todos los headers CORS necesarios.

- **A través de nginx**: `http://localhost` (única forma recomendada)
- **API directa**: No recomendado - no tendrá headers CORS

## Próximas Mejoras

- [ ] SSL/TLS con certificados
- [ ] Configuración de producción
- [ ] Monitoring con Prometheus
- [ ] Logs estructurados

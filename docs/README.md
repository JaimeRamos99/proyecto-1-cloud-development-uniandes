# Video Platform API Documentation

Esta carpeta contiene la documentación completa de la API de la plataforma de videos.

## 📖 Documentación OpenAPI/Swagger

### Archivos disponibles:

- `swagger.yaml` - Especificación completa de la API en formato OpenAPI 3.0.3

### 🌟 Características documentadas:

#### **🔐 Autenticación**

- Registro de usuarios con validación de email
- Login con JWT tokens
- Logout con invalidación de tokens
- Middleware de autenticación

#### **🎥 Videos**

- Upload de videos con validación de formato
- Gestión de videos privados y públicos
- Procesamiento asíncrono con worker
- URLs presignadas para acceso seguro
- Soft delete (solo videos privados)

#### **🗳️ Sistema de Votación**

- Votación en tiempo real
- Un voto por usuario por video
- Auto-actualización de rankings
- Mensajes personalizados y motivacionales

#### **🏆 Sistema de Rankings**

- Rankings dinámicos basados en votos
- Paginación y filtros avanzados
- Vista materializada para rendimiento
- Actualización automática tras cada voto
- Filtros por país, ciudad, cantidad de votos

#### **💊 Health Checks**

- Status de la API y dependencias
- Verificación de PostgreSQL y FFmpeg

### 🚀 Cómo visualizar la documentación:

#### **Opción 1: Swagger UI online**

1. Ve a [Swagger Editor](https://editor.swagger.io/)
2. Copia el contenido de `swagger.yaml`
3. Pégalo en el editor para visualizar interactivamente

#### **Opción 2: Localmente con Docker**

```bash
# Navega a la carpeta docs
cd docs

# Ejecuta Swagger UI con Docker
docker run -p 8080:8080 -v $(pwd):/usr/share/nginx/html -e SWAGGER_JSON=/usr/share/nginx/html/swagger.yaml swaggerapi/swagger-ui

# Abre http://localhost:8080 en tu navegador
```

#### **Opción 3: Usando el Makefile**

```bash
# Desde la raíz del proyecto
make docs

# Abre http://localhost:8080 en tu navegador
```

### 🔧 Estructura de la API

```
/api
├── /health                           # Health check
├── /auth
│   ├── POST /signup                 # Registro
│   ├── POST /login                  # Login
│   └── POST /logout                 # Logout
├── /videos                          # Videos privados (requiere auth)
│   ├── POST /upload                 # Upload video
│   ├── GET /                        # Listar mis videos
│   ├── GET /:video_id              # Detalle video
│   └── DELETE /:video_id           # Eliminar video
└── /public
    ├── GET /videos                  # Videos públicos
    ├── /videos/:video_id/vote
    │   ├── POST                     # Votar video
    │   └── DELETE                   # Quitar voto
    └── /rankings
        ├── GET /                    # Listar rankings
        ├── GET /:user_id           # Ranking específico
        └── POST /refresh           # Refresh manual
```

### 🎯 Ejemplos de uso

#### **Registro e Login**

```bash
# Registro
curl -X POST http://localhost:80/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "password1": "secretpass",
    "password2": "secretpass",
    "city": "Bogotá",
    "country": "Colombia"
  }'

# Login
curl -X POST http://localhost:80/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secretpass"
  }'
```

#### **Upload de video**

```bash
curl -X POST http://localhost:80/api/videos/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "video_file=@path/to/video.mp4" \
  -F "title=My Amazing Video" \
  -F "is_public=true"
```

#### **Obtener rankings**

```bash
# Rankings paginados con filtros
curl "http://localhost:80/api/public/rankings?page=1&page_size=10&country=Colombia&min_votes=5"

# Ranking específico
curl "http://localhost:80/api/public/rankings/123"
```

#### **Votar por video**

```bash
# Votar
curl -X POST http://localhost:80/api/public/videos/456/vote \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Quitar voto
curl -X DELETE http://localhost:80/api/public/videos/456/vote \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 📋 Códigos de respuesta comunes

| Código | Descripción                              |
| ------ | ---------------------------------------- |
| `200`  | Operación exitosa                        |
| `201`  | Recurso creado exitosamente              |
| `204`  | Operación exitosa sin contenido          |
| `400`  | Request inválido o datos mal formateados |
| `401`  | Autenticación requerida o token inválido |
| `403`  | Acceso denegado                          |
| `404`  | Recurso no encontrado                    |
| `409`  | Conflicto (ej: email ya existe, ya votó) |
| `500`  | Error interno del servidor               |

### 🔒 Autenticación

La API usa **JWT (JSON Web Tokens)** para autenticación:

1. Obtén un token con `POST /api/auth/login`
2. Incluye el token en el header: `Authorization: Bearer YOUR_TOKEN`
3. Los endpoints que requieren auth están marcados con 🔒 en la documentación

### 📊 Filtros y paginación

#### **Rankings - Parámetros disponibles:**

- `page` (int): Número de página (default: 1)
- `page_size` (int): Elementos por página (default: 10, max: 100)
- `country` (string): Filtrar por país
- `city` (string): Filtrar por ciudad
- `min_votes` (int): Votos mínimos
- `max_votes` (int): Votos máximos
- `min_videos` (int): Videos mínimos subidos
- `max_videos` (int): Videos máximos subidos

### 🎨 Respuestas de ejemplo

Todas las respuestas exitosas siguen estructuras consistentes. Los errores siempre retornan:

```json
{
  "error": "Descripción clara del error"
}
```

### 🔄 Tiempo real

Los rankings se actualizan **automáticamente** después de cada voto/unvote:

- No necesitas hacer polling
- Los cambios son instantáneos
- La vista materializada se refresca automáticamente

### 🚨 Limitaciones importantes

- **Videos públicos no se pueden eliminar** (para mantener integridad de rankings)
- **Un voto por usuario por video** (constraint de BD)
- **Upload máximo**: 100MB por video (configurado en nginx)
- **Formatos soportados**: MP4, AVI, MOV, MKV

---

## 📞 Soporte

- **Repository**: [GitHub](https://github.com/JaimeRamos99/proyecto-1-cloud-development-uniandes)
- **Issues**: [GitHub Issues](https://github.com/JaimeRamos99/proyecto-1-cloud-development-uniandes/issues)

---

**¡Documentación completa, API lista para usar! 🚀**

# Proyecto 1: Sistema de Gestión de Videos

## Integrantes

| Nombre                         | Correo                     |
| ------------------------------ | -------------------------- |
| Jaime Josue Ramos Rossetes     | jj.ramosr1@uniandes.edu.co |
| Marilyn Stephany Joven Fonseca | m.joven@uniandes.edu.co    |

## 🚀 Inicio Rápido

Este proyecto utiliza **Makefile** para simplificar el desarrollo. Se recomienda usar los comandos make en lugar de docker-compose directamente.

### Comandos Principales

```bash
# Ver todos los comandos disponibles
make help

# Iniciar todo el entorno local (recomendado)
make local

# Ver logs de todos los servicios
make logs

# Verificar estado de salud de servicios
make health

# Detener todos los contenedores
make stop

# Limpieza completa (elimina todo)
make clean
```

### Servicios Disponibles

Una vez ejecutado `make local`:

- **🌐 API**: http://localhost:80/api
- **📚 Documentación**: http://localhost:8080
- **🗄️ PostgreSQL**: localhost:5432
- **☁️ LocalStack**: http://localhost:4566

### Documentación

Toda la documentación del proyecto se encuentra en `docs/Entrega_1/`:

- **Diagrama ERD** - Modelo de datos y relaciones
- **Diagramas C4** - Arquitectura del sistema (niveles 1, 2 y 3)
- **Diagrama de flujo de procesamiento** - Proceso completo de carga y procesamiento de videos
- **Decisiones de Arquitectura** - Justificación de decisiones de diseño clave
- **Especificación API (Swagger)** - Documentación completa de endpoints

# Ejecución de pruebas 

Apache JMeter (5.6.3+)\
FFmpeg (para generar videos de prueba)\
Servidor desplegado en: http://13.223.138.92

```
📁 Estructura de Pruebas
pruebas-de-carga/
├── jmeter/
│   ├── test-plans/
│   │   ├── 01-funcional-basico.jmx       # Pruebas funcionales básicas
│   │   └── 02-carga-upload.jmx           # Pruebas de carga de uploads
│   └── data/
│       ├── usuarios.csv                   # Usuarios de prueba registrados
│       ├── videos.csv                     # Rutas de videos para upload
│       └── videos-prueba/                 # Videos generados (10MB-100MB)
├── scripts/
│   ├── generate-test-videos.sh           # Script para generar videos de prueba
│   ├── register-test-users.sh            # Script para registrar usuarios en servidor
│   ├── run-load-tests.sh                 # Script para ejecutae las pruebas
│   └── monitor-remote.sh                 # Monitorea servidor durante pruebas
├── reportes/                             # Reportes HTML generados
│   ├── funcional/
│   └── carga-normal/
└── README.md
```

### Configuración Inicial
1. Generación Videos de Prueba
bashcd pruebas-de-carga
Esto crea 5 videos de prueba (10MB, 25MB, 50MB, 75MB, 100MB) en formato MP4 1080p.
2. Registrar Usuarios de Prueba
Registra 20 usuarios de prueba en el servidor (testuser01@anb.com - testuser20@anb.com).

### Ejecutar Pruebas de Carga
Ejecución Automática de un script que se ejecuta secuencialmente:

* Prueba Funcional (5 usuarios, 2 min)
* Prueba de Carga Normal (10 usuarios, 5 min)

Los reportes HTML se generan automáticamente en reportes/.

### Configuración de las Pruebas
Prueba Funcional Básica\
Archivo: 01-funcional-basico.jmx\
Usuarios: 5 concurrentes\
Duración: ~2 minutos\
Objetivo: Verificar endpoints básicos (login, profile, videos públicos)\
Prueba de Carga - Upload de Videos\
Archivo: 02-carga-upload.jmx\
Usuarios: 10 concurrentes\
Ramp-up: 300 segundos (1 usuario cada 30s)\
Duración: ~10 minutos\

Flujo de cada usuario:

1. Login → Obtiene token JWT
2. Upload Video → Sube video de 10-100MB
3. Check Status → Verifica estado del procesamiento


### Monitoreo Durante Pruebas
Este script monitorea el servidor en tiempo real mientras se ejecutan las pruebas, mostrando:

✅ Conectividad: Latencia (ping) al servidor\
✅ Health Check: Estado de la API (/api/health)\
✅ Response Times: Tiempos de respuesta de endpoints clave:

/api/public/videos\
/api/public/rankings


🔄 Actualización cada 5 segundos

Ejemplo de salida:
=== Estado del Servidor ANB - 23:45:12 ===

[CONECTIVIDAD]
✓ 13.223.138.92 respondiendo (45ms)

[API HEALTH CHECK]
✓ API disponible (0.234s)

[ENDPOINTS DE PRUEBA]
  /api/public/videos: 0.456s
  /api/public/rankings: 0.289s

Próxima actualización en 5s...
Uso recomendado:

Abrir 2 terminales
Terminal 1: ./scripts/run-load-tests.sh (ejecuta pruebas)
Terminal 2: ./scripts/monitor-remote.sh (monitorea servidor)

Así puedes ver en tiempo real si el servidor está respondiendo mientras JMeter genera carga.

## Ver Reportes
Los reportes HTML se generan automáticamente en:
reportes/funcional/html_TIMESTAMP/index.html
reportes/carga-normal/html_TIMESTAMP/index.html
Abrir en navegador:
bashopen reportes/funcional/html_*/index.html


Métricas Clave a Revisar
MétricaObjetivoLímite CríticoResponse Time (API)< 500ms< 1000msUpload Time (50MB)< 60s< 120sThroughput> 50 req/min> 25 req/minError Rate< 2%< 5%

⚙️ Configuración de JMeter
Archivos CSV
usuarios.csv:
csvfirst_name,last_name,email,password,city,country
Juan,Pérez,testuser01@anb.com,Test123!,Bogotá,Colombia
...
videos.csv:
csvvideo_path
/Users/marilyn/Documents/.../video-10mb-20s-1080p.mp4
...
Endpoints Probados
EndpointMétodoAutenticación/api/auth/loginPOSTNo/api/auth/profileGETSí (Bearer)/api/videos/uploadPOSTSí (Bearer)/api/videos/:idGETSí (Bearer)/api/public/videosGETNo/api/public/rankingsGETNo
Datos de Upload
Cada request de upload envía:

video_file: Archivo MP4 (10-100MB)
title: "Test Video {timestamp}"
is_public: true


🐛 Troubleshooting
Problema: 502 Bad Gateway
Causa: Demasiados usuarios simultáneos saturan el servidor.
Solución: Reducir usuarios o aumentar ramp-up:
Threads: 5-10 (en lugar de 20)
Ramp-up: 300 segundos (en lugar de 60)
Problema: FileNotFoundException
Causa: Rutas incorrectas en videos.csv.
Solución: Regenerar CSV con rutas absolutas:
bash./scripts/generate-videos-csv.sh
Problema: 401 Unauthorized
Causa: Usuarios no registrados o tokens expirados.
Solución:
bash./scripts/register-test-users.sh


🔧 Ajustar Configuración
Para modificar las pruebas:

Abrir JMeter GUI:

bash   jmeter -t jmeter/test-plans/02-carga-upload.jmx

Ajustar Thread Group:

Number of Threads: Cantidad de usuarios
Ramp-up Period: Tiempo para iniciar todos los usuarios
Loop Count: Repeticiones por usuario


Guardar y ejecutar con el script.




Autor: Equipo ANB
Fecha: Octubre 2025
Servidor: http://13.223.138.92
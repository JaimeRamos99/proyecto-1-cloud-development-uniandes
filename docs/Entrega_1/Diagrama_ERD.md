# Diagrama Entidad-Relación (ERD) - Sistema de Videos

## Descripción General

Este documento presenta el modelo de datos de la aplicación de sistema de videos, que permite a los usuarios subir videos, votarlos y generar rankings basados en popularidad.

## Diagrama ERD

```
┌─────────────────────────────────────┐
│                USERS                │
├─────────────────────────────────────┤
│ PK  id: INTEGER (SERIAL)            │
│     first_name: TEXT                │
│     last_name: TEXT                 │
│ UQ  email: CITEXT                   │
│     password_hash: TEXT             │
│     city: TEXT                      │
│     country: TEXT                   │
└─────────────────────────────────────┘
                    │
                    │ 1:N
                    ▼
┌─────────────────────────────────────┐
│               VIDEOS                │
├─────────────────────────────────────┤
│ PK  id: INTEGER (SERIAL)            │
│     title: TEXT                     │
│     status: video_status            │
│     is_public: BOOLEAN              │
│     uploaded_at: TIMESTAMP          │
│     processed_at: TIMESTAMP?        │
│     deleted_at: TIMESTAMP?          │
│ FK  user_id: INTEGER                │
└─────────────────────────────────────┘
                    │
                    │ 1:N
                    ▼
┌─────────────────────────────────────┐
│                VOTES                │
├─────────────────────────────────────┤
│ PK  id: INTEGER (SERIAL)            │
│ FK  user_id: INTEGER                │
│ FK  video_id: INTEGER               │
│     voted_at: TIMESTAMP             │
│                                     │
│ UQ  (user_id, video_id)             │
└─────────────────────────────────────┘
                    ▲
                    │ N:1
                    │
┌─────────────────────────────────────┐
│                USERS                │
│     (referencia desde VOTES)        │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│          PLAYER_RANKINGS            │
│        (Materialized View)          │
├─────────────────────────────────────┤
│     user_id: INTEGER                │
│     first_name: TEXT                │
│     last_name: TEXT                 │
│     email: CITEXT                   │
│     city: TEXT                      │
│     country: TEXT                   │
│     total_votes: INTEGER            │
│     ranking: INTEGER                │
│     last_updated: TIMESTAMP         │
└─────────────────────────────────────┘
```

## Entidades y Atributos

### 1. USERS (Usuarios)

**Descripción**: Almacena la información de los usuarios registrados en el sistema.

| Atributo      | Tipo             | Descripción                           | Restricciones    |
| ------------- | ---------------- | ------------------------------------- | ---------------- |
| id            | INTEGER (SERIAL) | Identificador único del usuario       | PRIMARY KEY      |
| first_name    | TEXT             | Nombre del usuario                    | NOT NULL         |
| last_name     | TEXT             | Apellido del usuario                  | NOT NULL         |
| email         | CITEXT           | Correo electrónico (case-insensitive) | NOT NULL, UNIQUE |
| password_hash | TEXT             | Hash de la contraseña                 | NOT NULL         |
| city          | TEXT             | Ciudad del usuario                    | NOT NULL         |
| country       | TEXT             | País del usuario                      | NOT NULL         |

### 2. VIDEOS (Videos)

**Descripción**: Almacena la información de los videos subidos por los usuarios.

| Atributo     | Tipo             | Descripción                        | Restricciones                           |
| ------------ | ---------------- | ---------------------------------- | --------------------------------------- |
| id           | INTEGER (SERIAL) | Identificador único del video      | PRIMARY KEY                             |
| title        | TEXT             | Título del video                   | NOT NULL                                |
| status       | video_status     | Estado del video                   | NOT NULL, ENUM('uploaded', 'processed') |
| is_public    | BOOLEAN          | Visibilidad pública del video      | NOT NULL, DEFAULT false                 |
| uploaded_at  | TIMESTAMP        | Fecha de subida                    | NOT NULL, DEFAULT NOW()                 |
| processed_at | TIMESTAMP        | Fecha de procesamiento             | NULL                                    |
| deleted_at   | TIMESTAMP        | Fecha de eliminación (soft delete) | NULL                                    |
| user_id      | INTEGER          | Referencia al usuario propietario  | NOT NULL, FK → users(id)                |

### 3. VOTES (Votos)

**Descripción**: Almacena los votos que los usuarios dan a los videos.

| Atributo | Tipo             | Descripción                    | Restricciones             |
| -------- | ---------------- | ------------------------------ | ------------------------- |
| id       | INTEGER (SERIAL) | Identificador único del voto   | PRIMARY KEY               |
| user_id  | INTEGER          | Referencia al usuario que vota | NOT NULL, FK → users(id)  |
| video_id | INTEGER          | Referencia al video votado     | NOT NULL, FK → videos(id) |
| voted_at | TIMESTAMP        | Fecha del voto                 | NOT NULL, DEFAULT NOW()   |
|          |                  |                                | UNIQUE(user_id, video_id) |

### 4. PLAYER_RANKINGS (Vista Materializada)

**Descripción**: Vista materializada que calcula los rankings de usuarios basado en los votos recibidos.

| Atributo     | Tipo      | Descripción              | Notas |
| ------------ | --------- | ------------------------ | ----- |
| user_id      | INTEGER   | ID del usuario           |       |
| first_name   | TEXT      | Nombre del usuario       |       |
| last_name    | TEXT      | Apellido del usuario     |       |
| email        | CITEXT    | Email del usuario        |       |
| city         | TEXT      | Ciudad del usuario       |       |
| country      | TEXT      | País del usuario         |       |
| total_votes  | INTEGER   | Total de votos recibidos |       |
| ranking      | INTEGER   | Posición en el ranking   |       |
| last_updated | TIMESTAMP | Última actualización     |       |

## Relaciones

### 1. USERS → VIDEOS (1:N)

- **Cardinalidad**: Un usuario puede tener muchos videos
- **Relación**: `users.id` ← `videos.user_id`
- **Acción**: ON DELETE CASCADE

### 2. USERS → VOTES (1:N)

- **Cardinalidad**: Un usuario puede dar muchos votos
- **Relación**: `users.id` ← `votes.user_id`
- **Acción**: ON DELETE CASCADE

### 3. VIDEOS → VOTES (1:N)

- **Cardinalidad**: Un video puede recibir muchos votos
- **Relación**: `videos.id` ← `votes.video_id`
- **Acción**: ON DELETE CASCADE

### Restricción Única

- **VOTES**: `UNIQUE(user_id, video_id)` - Un usuario solo puede votar una vez por video

## Tipos de Datos Especiales

### ENUM video_status

```sql
CREATE TYPE video_status AS ENUM (
  'uploaded',    -- Video subido pero no procesado
  'processed'    -- Video procesado y listo
);
```

### CITEXT Extension

- Se utiliza para el campo `email` en la tabla `users`
- Permite comparaciones case-insensitive
- `'User@example.com'` y `'user@example.com'` se consideran iguales

## Índices Principales

### USERS

- PRIMARY KEY: `users_pkey` en `id`
- UNIQUE INDEX: en `email`

### VIDEOS

- PRIMARY KEY: `videos_pkey` en `id`
- INDEX: `idx_videos_user_id` en `user_id`
- INDEX: `idx_videos_is_public` en `is_public`
- INDEX: `idx_videos_user_id_is_public` en `(user_id, is_public)`

### VOTES

- PRIMARY KEY: `votes_pkey` en `id`
- INDEX: `idx_votes_user_id` en `user_id`
- INDEX: `idx_votes_video_id` en `video_id`
- INDEX: `idx_votes_user_video` en `(user_id, video_id)`

### PLAYER_RANKINGS (Vista Materializada)

- UNIQUE INDEX: `idx_player_rankings_user_id` en `user_id`
- INDEX: `idx_player_rankings_total_votes` en `total_votes DESC`
- INDEX: `idx_player_rankings_ranking` en `ranking`
- INDEX: `idx_player_rankings_country` en `country`
- INDEX: `idx_player_rankings_city` en `city`

## Funcionalidades del Modelo

1. **Gestión de Usuarios**: Registro y autenticación con emails case-insensitive
2. **Gestión de Videos**: Subida, procesamiento y visibilidad pública/privada
3. **Sistema de Votación**: Votos únicos por usuario-video
4. **Rankings Dinámicos**: Vista materializada que se actualiza automáticamente
5. **Soft Delete**: Los videos pueden ser marcados como eliminados sin perder datos
6. **Auditoría**: Timestamps de creación y procesamiento

## Consideraciones de Rendimiento

- **Índices Optimizados**: Para consultas frecuentes por usuario, estado y visibilidad
- **Vista Materializada**: Para rankings con actualización automática cada minuto
- **Cascadas**: Eliminación en cascada para mantener integridad referencial
- **Paginación**: Soporte para consultas paginadas en rankings

## Integridad Referencial

- Todas las foreign keys tienen `ON DELETE CASCADE`
- Constraint único en votos previene duplicados
- Extension CITEXT garantiza unicidad de emails independiente de mayúsculas

# Proyecto 1: Sistema de Gestión de Videos

## Integrantes

| Nombre                         | Correo                     |
| ------------------------------ | -------------------------- |
| Jaime Josue Ramos Rossetes     | jj.ramosr1@uniandes.edu.co |
| Marilyn Stephany Joven Fonseca | m.joven@uniandes.edu.co    |


Este proyecto presenta una solución cloud-native diseñada para subir, procesar y publicar videos de manera eficiente, escalable y segura.

Los reportes incluidos documentan la transición completa entre versiones, detallando diseño, decisiones técnicas, despliegues y mejoras progresivas. En este README se encuentran enlaces a cada entrega de arquitectura, así como la documentación de la API implementada.

El proyecto se construyó siguiendo principios del AWS Well-Architected Framework y adoptando prácticas modernas como separación de responsabilidades, escalamiento automático, infraestructura como código, uso de contenedores, procesamiento asíncrono y entrega global mediante CDN.

En conjunto, los reportes explican cómo se implementaron los siguientes pilares:

- Separación total entre frontend y backend
- Migración del frontend a S3 + CloudFront
- Backend distribuido detrás de un Application Load Balancer
- Auto Scaling Group para alta disponibilidad y resiliencia
- Contenedores almacenados en ECR para despliegues reproducibles
- Worker evolucionando desde EC2 tradicional hacia Lambda serverless
- Procesamiento asíncrono basado en SQS
- Presigned URLs para subir videos directamente a S3
- Seguridad reforzada con IAM, SGs y VPC endpoints
- Infraestructura declarativa en Terraform
- Pipeline CI/CD completo para API, worker y frontend
- Monitoreo centralizado con CloudWatch
- Organización avanzada del almacenamiento en S3
- Optimización de costos mediante lifecycle policies
- Base de datos RDS con backups automáticos
- Worker con paralelización y tiempos de ejecución controlados


### Documentación

Toda la documentación para la Entrega 1 del proyecto se encuentra en `docs/Entrega_1/`:

- **Diagrama ERD** - Modelo de datos y relaciones
- **Diagramas C4** - Arquitectura del sistema (niveles 1, 2 y 3)
- **Diagrama de flujo de procesamiento** - Proceso completo de carga y procesamiento de videos
- **Decisiones de Arquitectura** - Justificación de decisiones de diseño clave
- **Especificación API (Swagger)** - Documentación completa de endpoints

Toda la documentación para la Entrega 2 del proyecto se encuentra en `docs/Entrega_2/`:

- **Diagrama de Arquitectura** - Arquitectura definida
- **Diagrama de Despliegue** - Diagrama de despliegue
- **Documentacion Arquitectura** - Explicación del diseño de la arquitectura.
- **Documentacion Pruebas** - Procedimiento y explicación de la estructura de las pruebas de carga.
- **Pruebas de carga** - Ejecución de las pruebas de carga
- **Reportes PDF** - Resumen de los reportes de las pruebas de Funcionalidad, Carga Normal y Estrés.


Toda la documentación para la Entrega 3 del proyecto se encuentra en `docs/Entrega_3/`:

- **Documentacion Arquitectura** - Explicación del diseño de la arquitectura.
- **Documentacion Pruebas** - Procedimiento y explicación de la estructura de las pruebas de carga.
- **Pruebas de carga** - Ejecución de las pruebas de carga
- **Reportes PDF** - Resumen de los reportes de las pruebas de Funcionalidad, Carga Normal y Estrés.



Toda la documentación para la Entrega 4 y 5 del proyecto se encuentra en `docs/Entrega_4_5/`:

- **Diagrama de Arquitectura** - Arquitectura final
- **Documentacion Arquitectura** - Explicación del diseño de la arquitectura.
- **Documentacion Pruebas** - Procedimiento y explicación de la estructura de las pruebas de carga.
- **Pruebas de carga** - Ejecución de las pruebas de carga
- **Reportes PDF** - Resumen de los reportes de las pruebas de Funcionalidad, Carga Normal y Estrés.
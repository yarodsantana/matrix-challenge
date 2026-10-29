# Decisiones de arquitectura y cumplimiento

## 1. Responsabilidades exigidas

Se mantienen dos APIs RESTful independientes, una en Go/Fiber y otra en Node/Express. Go recibe la matriz original, realiza las transformaciones y envia resultados a Node por HTTP. Node realiza el calculo adicional y devuelve estadisticas. No se trasladan las estadisticas principales al frontend ni se replica su calculo productivo en Go.

El navegador utiliza Go como punto de entrada. Consumir Node directamente permitiria saltarse la transformacion y obligaria a entregar una credencial de servicio al navegador. Por eso el frontend interactua con ambos servicios mediante la orquestacion requerida, conservando la direccion de comunicacion del documento.

## 2. Ambiguedad rotacion / QR

El requerimiento funcional explicito pide QR de la matriz recibida, mientras que la descripcion de arquitectura habla de rotarla. Decision informada: /qr factoriza la original A, devuelve su rotacion por separado y envia exclusivamente Q y R a Node. La rotacion no reemplaza QR ni se factoriza silenciosamente la matriz rotada.

/rotate se conserva para evitar romper el contrato inicial; el frontend distingue ambos modos. Esta decision cubre las dos operaciones y permite explicar su diferencia durante la entrevista.

## 3. QR completa y estabilidad

Se usa QR completa: Q de m x m, R de m x n. Esto funciona para cualquier matriz rectangular, incluso ancha o de rango deficiente, sin restricciones ocultas m>=n. Las reflexiones de Householder son preferibles a Gram-Schmidt clasico por su estabilidad. Se implementa el algoritmo en un paquete matematico aislado, con comentarios sobre signo, escalado y acumulacion.

No se impone que la diagonal de R sea positiva. QR admite distintas convenciones de signo; las estadisticas dependen de esa convencion, que queda declarada. El criterio matematico de aceptacion es A aproximadamente QR y Q^T Q aproximadamente I, no comparar una representacion de Q/R unica.

Se limita QR a 100x100: Q completa ocupa O(m^2), R ocupa O(mn) y acumular Q tiene coste proporcional a m^2 min(m,n). Ese limite permite una demostracion interactiva controlada sin introducir colas o infraestructura adicional.

Referencias del algoritmo y convencion completa:
- https://netlib.org/lapack/explore-html/d0/da1/group__geqrf_gade26961283814bb4e62183d9133d8bf5.html
- https://www.netlib.org/lapack/lawnspdf/lawn178.pdf

## 4. Estadisticas y diagonal

Node agrega todos los valores de Q y R. Incluye los ceros, incluidos los de relleno de R completa. Informa max, min, sum, average y si alguna matriz es diagonal. Ademas expone perMatrix y diagonalMatrices para que el resultado pueda auditarse.

Una diagonal es cuadrada y tiene ceros fuera de la diagonal principal. No debe confundirse con triangular superior. En QR se aplica tolerancia 1e-10*max(1,maxAbs) por matriz para aceptar ruido de punto flotante. En /rotate se conserva la semantica exacta previa.

## 5. Nombres coherentes

El enunciado no impone nombres pero si coherencia. Se usan rutas por operacion (/qr, /rotate), q/r para las matrices matematicas, original/rotated para distinguir entrada y rotacion, statistics para el analisis, y camelCase en propiedades compuestas (isDiagonal, valueCount, diagonalMatrices, perMatrix). No se mezclan distintos nombres para el mismo campo.

Los paquetes Go tienen nombres cortos por responsabilidad. Las funciones publicas tienen comentarios de contrato y las decisiones no obvias tienen explicacion; se evitan comentarios que repitan cada instruccion. Los errores HTTP se manejan de forma compartida, sin duplicar el calculo.

## 6. Seguridad y frontend

JWT HS256 con algoritmo, emisor, audiencia y vencimiento exigidos. Claves distintas para cliente-Go y Go-Node. Configuracion por entorno, sin defaults secretos. /health y assets son publicos; /qr y /rotate requieren JWT.

El frontend mantiene el token solo en memoria, renderiza valores con textContent, no evalua HTML del usuario y no incluye dependencias de CDN. CSP limita scripts y conexiones al mismo origen. Se puede copiar/descargar el JSON de resultados; los tokens no forman parte de ese JSON.

El emisor local es una herramienta del desarrollador, no un login abierto. No se ha inventado un sistema de usuarios que el documento no solicita.

## 7. Docker y Render Free

Se construyen y prueban cada API, se ejecutan como usuarios sin root y se excluyen caches/secretos del contexto Docker. El contenedor Go incorpora el frontend y certificados para la llamada HTTPS a Node en la nube.

Local: Node permanece interno y Go publica un puerto. Render Free: dos Web Services publicos porque Private Services no ofrece plan Free. JWT protege Node aunque su URL sea publica. Se usa un timeout de 90s y tokens de servicio de 2min para el arranque en frio. Las claves solo se inyectan en runtime.

No se crean recursos de pago ni se configuran pings para evitar la suspension de servicios gratuitos.

## 8. Trazabilidad de requisitos

| Requisito | Implementacion / evidencia |
| --- | --- |
| Go + Fiber | api-go/cmd/api |
| Node + Express | api-node/src/app.js |
| Matriz rectangular | Validate y readMatrix |
| Factorizacion QR | matrix.FactorQR, propiedades algebraicas y pruebas aleatorias |
| Rotacion | RotateClockwise, expuesta tambien en /qr |
| HTTP entre APIs | statistics.Client + integracion con Express real |
| Estadisticas sobre matrices | calculateQRStatistics |
| Alguna diagonal | isDiagonal agregado + diagonalMatrices |
| Docker | dos Dockerfiles + compose.yaml |
| Nube | render.yaml y guia; evidencia de URLs pendiente hasta desplegar |
| Frontend opcional | frontend/, servido por Go |
| JWT opcional | auth en ambas APIs y pruebas de rechazo |
| Pruebas opcionales | unitarias, HTTP, algebraicas e integracion |
| Documentacion / IA | READMEs, docs y PROMPTS.md |
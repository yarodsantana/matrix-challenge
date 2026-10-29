# Autenticación JWT

Se protegen POST /qr y POST /rotate en Go, y POST /statistics y POST /statistics/qr en Node. GET /health y los assets del frontend son publicos.

## Relaciones de confianza

| Campo | Cliente a Go | Go a Node |
| --- | --- | --- |
| Algoritmo permitido | HS256 | HS256 |
| Clave de firma | JWT_SECRET | SERVICE_JWT_SECRET |
| Emisor (iss) | matrix-challenge | matrix-api-go |
| Destinatario (aud) | matrix-api-go | matrix-api-node |
| Sujeto (sub) | No vacío; technical-test-user en la herramienta local | matrix-api-go |
| Duración emitida | 15 minutos | 2 minutos |
| Vencimiento (exp) | Obligatorio | Obligatorio |

La firma autentica los claims; no cifra su contenido. Se verifican firma, algoritmo, emisor, audiencia y vencimiento antes de procesar la matriz. nbf e iat se validan cuando existen. Se permite una tolerancia de reloj explicita de 5 segundos para exp/nbf/iat, debido al pequeño desfase observado entre Windows y Render. Fechas fuera de ese margen se rechazan.

Un token ausente o inválido devuelve HTTP 401:

    {"error":"token ausente o invalido"}

La respuesta incluye WWW-Authenticate: Bearer. No se revelan al cliente detalles internos de la verificacion.

## Claves y configuración

- Claves distintas de al menos 32 bytes cada una.
- Sin valores secretos por defecto: el servicio no inicia con configuración inválida.
- Go recibe ambas claves; Node recibe solo SERVICE_JWT_SECRET.
- Las claves se inyectan por entorno al ejecutar los contenedores, sin incluirlas en las imagenes.
- .env esta excluido de Git y Docker. .env.example contiene valores vacios, no credenciales utilizables.
- Los tests usan claves publicas identificadas como test-only, independientes del .env real.

Desde la raíz, crear .env una sola vez:

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\init-env.ps1

El script conserva un .env existente. En Render se copian las mismas claves desde el archivo local a las variables de los servicios, sin pegarlas en logs o chats.

## Obtener un token para probar

Con Node y las dependencias locales de api-node instaladas:

    $token = node --env-file=.env .\scripts\create-token.mjs
    $token | Set-Clipboard

Pegar el token en el frontend. La aplicación lo mantiene solo en memoria; no lo guarda en localStorage ni lo devuelve dentro del JSON de resultados.

La herramienta de emisión es local y requiere la clave del desarrollador. No hay un endpoint publico que emita tokens sin verificar identidad; tampoco se implementa una base de usuarios o un login fuera del alcance solicitado.

## Flujo entre servicios

Go valida el token del cliente, procesa la matriz y firma un token NUEVO destinado a Node. No reenvia la credencial del cliente. Cada llamada interna obtiene un token de servicio con dos minutos de duración, suficiente para el timeout de 90 segundos usado durante el arranque en frio de Render Free.

Una respuesta 401 de Node se trata como fallo del servicio remoto y Go devuelve 502 al cliente; no confunde ese fallo con el JWT del usuario. Un timeout devuelve 504.

## Validación

Las pruebas comprueban tokens válidos, ausentes, mal formados, firma incorrecta, algoritmo no permitido, falta de exp, emisor/audiencia/sujeto incorrectos y fechas fuera del margen. Un adelanto de 3 segundos se admite; adelantos y vencimientos de 60 segundos se rechazan.

La integración verifica token emitido desde Node -> validación en Go -> token de servicio firmado por Go -> validación en Express. La evidencia publica de ejecución esta en VALIDATION.md.

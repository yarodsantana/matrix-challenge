# Despliegue en Render Free

Repositorio: https://github.com/yarodsantana/matrix-challenge

Se mantienen dos APIs y dos contenedores. En Docker local, Node esta en una red interna. En Render Free se usan DOS Web Services publicos, porque los Private Services no tienen plan gratuito. Node sigue protegido por JWT de servicio; el navegador solo consume Go. No se publica ninguna clave en el frontend.

## 1. Publicar el código en GitHub

La carpeta local ya esta inicializada con Git y origin apunta al repositorio indicado. .env y api-go/work estan ignorados. Desde la raíz, después del commit preparado:

    git push -u origin main

Si Git solicita autenticación, completa el inicio de sesion de GitHub desde tu terminal. No pegues tokens o contraseñas en el chat. Si git no esta en PATH, el script scripts/publish.ps1 intenta localizar también el Git incluido con Codex.

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\publish.ps1

## 2. Crear Node primero

En https://dashboard.render.com selecciona New > Web Service, conecta GitHub y elige matrix-challenge.

| Campo | Valor |
| --- | --- |
| Name | api-node-matrix-challenge (o un nombre disponible) |
| Branch | main |
| Language / Runtime | Docker |
| Root Directory | dejar vacío |
| Dockerfile Path | api-node/Dockerfile |
| Docker Build Context | . (raíz del repositorio) |
| Instance Type | Free |
| Health Check Path | /health |

Variables de entorno:
- PORT=3000
- NODE_ENV=production
- SERVICE_JWT_SECRET: copiar exactamente el valor del .env local, sin comillas.

No copiar JWT_SECRET a Node. Esperar estado Live. Abrir https://matrix-challenge-node.onrender.com/health y comprobar {"status":"ok"}. Guardar esa URL real (Render puede agregar un sufijo).

## 3. Crear Go + frontend

Crear un segundo Web Service desde el mismo repositorio y branch.

| Campo | Valor |
| --- | --- |
| Name | api-go-matrix-challenge (o un nombre disponible) |
| Runtime | Docker |
| Root Directory | dejar vacío |
| Dockerfile Path | api-go/Dockerfile |
| Docker Build Context | . |
| Instance Type | Free |
| Health Check Path | /health |

Variables de entorno:
- PORT=8080
- FRONTEND_DIR=/app/frontend
- STATISTICS_URL=https://matrix-challenge-node.onrender.com/statistics
- STATISTICS_TIMEOUT_SECONDS=90
- JWT_SECRET: valor del .env local.
- SERVICE_JWT_SECRET: mismo valor que en Node.

El Dockerfile Go necesita el contexto raíz porque incorpora frontend/. Dejar Root Directory vacío; no usar api-go como contexto.

Esperar Live y abrir la URL publica de Go. La pagina inicial es el frontend. La comunicación Go->Node usa HTTPS y los certificados de la imagen final.

## 4. Probar

Desde la raíz local:

    $token = node --env-file=.env .\scripts\create-token.mjs
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1 -BaseUrl 'https://api-go-matrix-challenge.onrender.com' -Token $token

Para pegar el token en la interfaz sin mostrarlo en la consola:

    $token | Set-Clipboard

Abrir la URL de Go, pegar en Token, ingresar la matriz y calcular. El token de cliente vence en 15 minutos; generar otro cuando corresponda.

Antes de la entrevista, visitar /health en Node y en Go para que ambos servicios esten activos. Render Free duerme servicios inactivos y puede tardar aproximadamente un minuto en despertarlos. No se han configurado pings artificiales ni servicios de pago. La primera consulta puede requerir reintento. El cliente Go admite 90 segundos en Render; por defecto local son 3 segundos. El token de servicio dura dos minutos para cubrir ese margen.

## Alternativa: Blueprint

render.yaml define los dos servicios Free. En New > Blueprint, seleccionar el repositorio e indicar las variables marcadas sync:false. SERVICE_JWT_SECRET se copia de Node a Go con fromService. Como la URL publica de Node se conoce después de crearlo, puede usarse temporalmente https://node-service.invalid/statistics y reemplazarse en Go al conocer la URL real. Ese dominio no es una direccion de despliegue.

No seleccionar una instancia de pago. Esta configuración no incluye bases de datos, discos ni servicios privados.

## Evidencia de despliegue

Completar estos campos solo después de verificar los servicios reales:
- URL Go/frontend: https://api-go-matrix-challenge.onrender.com (Live; salud, frontend y recorrido completo verificados).
- URL Node: https://matrix-challenge-node.onrender.com (Live, commit 3b6b868; salud y estadísticas verificadas).
- Primera versión remota validada: 3b6b868. La revision final con tolerancia JWT queda pendiente del envio final.
- Resultado del smoke test remoto: aprobado el 29 de octubre de 2026 para rotación, QR, JWT y frontend.

No se declara cumplido el requisito de nube hasta tener las dos URLs reales y una prueba remota aprobada.

Referencias oficiales:
- https://render.com/docs/docker
- https://render.com/docs/blueprint-spec
- https://render.com/docs/free

### Error /test not found
Ambos Dockerfiles usan ahora contexto raíz. Root Directory debe quedar vacío; Dockerfile Path es api-node/Dockerfile o api-go/Dockerfile. Node copia api-node/package*.json, api-node/src y api-node/test. Este ajuste resuelve el error de checksum cuando Render usa su contexto raíz predeterminado.

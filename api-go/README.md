# API Go / Fiber

El punto de entrada es cmd/api. main.go configura puerto, cliente HTTP, autenticacion y rutas; handlers.go valida la entrada y coordina los servicios.

- POST /qr: entrada matriz original, respuesta original/rotated/q/r/statistics/metadata.
- POST /rotate: conserva respuesta matrix/statistics de la implementacion inicial.
- GET /health: publico para Docker/Render.
- GET /, /app.js y /styles.css: frontend publico, sin claves incorporadas.

QR y rotacion viven en internal/matrix. El cliente de Node en internal/statistics comparte transporte y mantiene separados los contratos legacy y QR. JWT en internal/auth.

Desde esta carpeta:

    go test ./... -count=1
    go test -tags integration ./... -count=1
    go vet ./...

Las pruebas de integracion necesitan Node y npm ci en ../api-node. Si Node no esta en PATH, configurar NODE_BINARY. El frontend de pruebas se localiza en ../frontend a partir del repositorio. El Dockerfile Go se construye con contexto RAIZ del repositorio, no con api-go, porque incorpora frontend/.

Para ejecutar localmente, cargar las claves sin imprimirlas desde la raiz:

    Get-Content .env | ForEach-Object { if ($_ -match '^(JWT_SECRET|SERVICE_JWT_SECRET)=(.+)$') { [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process') } }
    cd api-go
    go run ./cmd/api

Node debe estar ejecutandose en el puerto 3000, o configurar STATISTICS_URL. Go usa PORT=8080 por defecto, STATISTICS_TIMEOUT_SECONDS=3 y FRONTEND_DIR=../frontend. Configuracion y ejemplos actuales: ../README.md. Justificacion: ../docs/ARCHITECTURE.md.
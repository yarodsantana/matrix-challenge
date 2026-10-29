# API Node / Express

- POST /statistics recibe una matriz y conserva el contrato original.
- POST /statistics/qr recibe {q,r} y agrega todos los valores de ambas matrices.
- GET /health es publico; ambos endpoints POST requieren JWT de servicio.

src/statistics.js contiene funciones puras; src/app.js configura middleware y errores; src/auth.js verifica JWT; src/server.js abre el puerto.

Desde esta carpeta:

    npm.cmd ci
    npm.cmd test
    node --env-file=../.env src/server.js

Se requiere SERVICE_JWT_SECRET. PORT es 3000 por defecto. No se necesitan JWT_SECRET ni tokens de cliente en esta API.

El promedio de Q/R se calcula con la cantidad TOTAL de elementos, sin promediar los promedios. isDiagonal agregado indica si alguna es diagonal; diagonalMatrices enumera q/r y perMatrix permite inspeccionar cada resultado. QR usa tolerancia por matriz 1e-10*max(1,maxAbs); legacy mantiene cero exacto.

La suite actual reporta 67 pruebas aprobadas: calculos, validaciones, JWT, HTTP y agregacion QR. El Dockerfile ejecuta las pruebas al construir. Todos los ejemplos y despliegue: ../README.md y ../docs/DEPLOY_RENDER.md.
El Dockerfile usa contexto RAIZ del repositorio. Construccion manual: docker build -f api-node/Dockerfile . desde la raiz. En Render, dejar Root Directory vacio.

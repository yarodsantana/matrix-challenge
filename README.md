# Matrix Challenge — Go + Fiber / Node.js + Express

[Arquitectura del proyecto](docs/PROJECT_ARCHITECTURE.md) · [Contratos HTTP](docs/API.md) · [Despliegue](docs/DEPLOY_RENDER.md)

Dos APIs REST, comunicación HTTP, factorización QR, rotación, estadísticas, frontend interactivo, JWT y Docker. Código y pruebas preparados para Render Free.

## Inicio rapido con Docker

Desde la raíz del repositorio:

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\init-env.ps1
    docker compose up --build --wait --wait-timeout 120
    docker compose ps
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1

Abrir http://localhost:8080. En Generar token, escribir la contraseña compartida con el evaluador. El token se coloca automáticamente en el formulario y dura 15 minutos. Después ingresar la matriz, seleccionar la operación y calcular.

El operador debe configurar DEMO_PASSWORD en .env (mínimo 12 caracteres). La contraseña real no se incluye en el repositorio ni en el JavaScript. Sin esta variable, la emisión devuelve 503; los tokens existentes y las APIs conservan su funcionamiento.

La herramienta scripts/create-token.mjs se mantiene como alternativa operativa para pruebas y automatización, sin ser necesaria para el uso normal del frontend.
Si el puerto 8080 ya esta ocupado:

    $env:GO_PORT='8081'
    docker compose up --build --wait --wait-timeout 120
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1 -BaseUrl http://localhost:8081

Para detener: docker compose down. Para ver registros: docker compose logs --tail 100.

## Flujo entre componentes

    Navegador / cliente
            | HTTP + JWT cliente
            v
       Go + Fiber  ------ HTTP + JWT servicio ------> Node + Express
       valida A                                       calcula estadisticas
       factoriza A=QR                                 sobre Q y R
       rota A por separado                            verifica alguna diagonal
       sirve frontend                                 devuelve resultado
            |<---------------- estadisticas ----------------|
            v
       respuesta JSON al cliente

Se respeta la direccion Go -> Node del desafio. El frontend consume Go y obtiene resultados de ambos servicios a traves de ese flujo. No accede directamente a Node ni necesita un token interno. Compartir origen evita CORS y un tercer servidor.

En Docker, Node solo es accesible en la red Compose. En Render Free, ambos son Web Services con URL publica; Node exige JWT de servicio. Arquitectura por capas, módulos y despliegue: docs/PROJECT_ARCHITECTURE.md. Decisiones funcionales y matemáticas: docs/ARCHITECTURE.md.

## Interpretacion del enunciado

El documento mezcla rotación y factorización QR. Son operaciones distintas. Se implementan ambas y se explicita su relacion:

- POST /qr: calcula QR de la matriz ORIGINAL A, devuelve también la rotación de A y envia Q/R a Node para estadísticas agregadas.
- POST /rotate: conserva el contrato inicial de rotación y estadísticas sobre la matriz rotada. No se elimina ni se cambia su respuesta.
- QR completa: para entrada m x n, Q es m x m y R es m x n.
- La rotación independiente es de 90 grados en sentido horario.
- Las estadísticas principales de /qr incluyen TODOS los elementos de Q y R, incluidos ceros.
- isDiagonal en /qr significa que AL MENOS UNA de Q o R es diagonal. Se informa cuales y se devuelven también estadísticas por matriz.

No hay ambiguedad oculta ni se considera una rotación como sustituto de QR. La eleccion de nombres es coherente (camelCase en metadatos JSON, q/r por convencion matematica) y esta documentada.

## Contratos REST

| Servicio | Método / ruta | Entrada | Autenticación |
| --- | --- | --- | --- |
| Go | GET / | Frontend | Publica |
| Go / Node | GET /health | Ninguna | Publica |
| Go | POST /auth/token | Objeto {password} | Contraseña de demostración |
| Go | POST /qr | Array de arrays numéricos | JWT cliente |
| Go | POST /rotate | Array de arrays numéricos | JWT cliente |
| Node | POST /statistics/qr | Objeto {q, r} | JWT servicio |
| Node | POST /statistics | Array de arrays numéricos | JWT servicio |

Ejemplo principal (desde la raíz):

    $token = node --env-file=.env .\scripts\create-token.mjs
    Invoke-RestMethod -Method Post -Uri http://localhost:8080/qr -Headers @{Authorization="Bearer $token"} -ContentType application/json -Body '[[1,2,3],[4,5,6]]' | ConvertTo-Json -Depth 12

Ejemplo numérico completo para entrada [[2]]:

    {
      "original": [[2]],
      "rotated": [[2]],
      "q": [[-1]],
      "r": [[-2]],
      "statistics": {
        "max": -1, "min": -2, "average": -1.5, "sum": -3,
        "isDiagonal": true, "matrixCount": 2, "valueCount": 2,
        "diagonalMatrices": ["q", "r"], "diagonalTolerance": 1e-10,
        "perMatrix": {
          "q": {"max":-1,"min":-1,"average":-1,"sum":-1,"isDiagonal":true},
          "r": {"max":-2,"min":-2,"average":-2,"sum":-2,"isDiagonal":true}
        }
      },
      "metadata": {"factorization":"A = Q R","qrMode":"full","rotationDegrees":90,"rotationDirection":"clockwise"}
    }

Q y R pueden contener signos negativos aunque A sea positiva; (-1)*(-2)=2. El algoritmo y la convencion estan documentados.
Errores JSON con propiedad error: 400 entrada inválida; 401 JWT ausente/inválido; 413 cuerpo excesivo; 415 Content-Type no admitido; 422 desbordamiento numérico; 502 respuesta inválida o fallo de Node; 504 timeout de Node.

## Matemáticas y límites

Householder calcula A=QR sin exigir rango completo. Se admiten matrices altas, anchas, cuadradas, nulas, columnas dependientes y negativas/decimales. Q/R no son unicas: se usa la convencion de signos producida por las reflexiones, sin forzar diagonal positiva en R.

QR permite hasta 100 filas y 100 columnas para limitar CPU, memoria y tamaño de respuesta. El cuerpo HTTP se limita a 1 MiB. Se rechazan matrices vacias, irregulares, null, texto y numeros no finitos. Se escala A para calcular normas y se rechaza un resultado que exceda float64.

Node usa Number, Go float64: no se promete aritmetica decimal exacta. En QR, una matriz cuadrada se considera diagonal cuando cada elemento fuera de la diagonal cumple abs(valor) <= 1e-10 * max(1, maxAbsDeLaMatriz). El endpoint legacy conserva comparacion exacta con cero.

Promedio agregado = (suma(Q) + suma(R)) / (cantidad(Q) + cantidad(R)). No es el promedio simple de los promedios cuando las matrices tienen distinto tamaño.

## Carpetas

    api-go/cmd/api/              configuracion, endpoints y pruebas HTTP/integracion
    api-go/internal/matrix/      validacion, rotacion y QR Householder
    api-go/internal/statistics/  cliente HTTP autenticado de Node
    api-go/internal/auth/        JWT cliente/servicio
    api-go/internal/web/         rutas conocidas de assets
    api-node/src/               Express, JWT y calculos estadisticos
    api-node/test/              pruebas matematicas y HTTP
    frontend/                   HTML, Bootstrap y jQuery incluidos localmente
    scripts/                    configuracion, JWT, smoke test y publicacion Git
    docs/                       arquitectura, JWT y despliegue

## Variables de entorno

| Variable | Servicio | Uso |
| --- | --- | --- |
| DEMO_PASSWORD | Go | Contraseña de acceso de demostración; mínimo 12 caracteres |
| JWT_SECRET | Go / herramienta local | Clave de cliente, mínimo 32 bytes |
| SERVICE_JWT_SECRET | Go y Node | Otra clave distinta, mínimo 32 bytes |
| STATISTICS_URL | Go | URL de Node terminada en /statistics; /qr se deriva para estadísticas QR |
| STATISTICS_TIMEOUT_SECONDS | Go | 3 por defecto; 90 en Render Free; rango 1..120 |
| PORT | Ambos | 8080 en Go y 3000 en Node por defecto |
| FRONTEND_DIR | Go | ../frontend local; /app/frontend en Docker |
| GO_PORT | Compose | Puerto del equipo publicado para Go |

.env no se sube a Git ni se incluye en imagenes. Los tokens cliente duran 15 minutos; los de servicio, 2 minutos para cubrir el arranque en frio de Render Free. Consultar docs/JWT.md.

## Pruebas

Desde api-node:

    npm.cmd ci
    npm.cmd test

Desde api-go:

    go test ./... -count=1
    go test -tags integration ./... -count=1
    go vet ./...

Integración requiere Node y dependencias instaladas. Si node no esta en PATH, configurar NODE_BINARY. En Windows también se ejecuta el smoke test contra ambos servicios reales en puertos temporales. Los Dockerfiles ejecutan pruebas antes de producir las imagenes finales.

Cobertura de comportamiento: A aproximadamente igual a QR, Q^T Q aproximadamente identidad, R triangular superior, 64 matrices aleatorias deterministas, matrices altas/anchas/deficientes, validaciones, estadísticas de Q+R, JWT, errores HTTP, timeout y preservacion de /rotate. El frontend se verifica como assets servidos y por sintaxis; no se afirma una prueba visual automatizada de navegador.

## Nube y evidencia

Guia paso a paso: docs/DEPLOY_RENDER.md. Blueprint opcional: render.yaml. Ambos servicios se configuran como Free, sin bases de datos ni discos.

Ambos servicios estan Live y el smoke test remoto del recorrido completo paso el 29 de octubre de 2026. La ultima revision de tolerancia de reloj debe desplegarse con el envio final del código. La evidencia publica esta en docs/VALIDATION.md; las decisiones tecnicas se sustentan en docs/ARCHITECTURE.md. El registro personal de prompts se conserva fuera del repositorio y de Docker.

Tolerancia de reloj JWT: 5 segundos para exp, nbf e iat. Se observo un pequeño desfase real entre Windows y Render; esta ventana evita rechazar tokens recien emitidos y sigue rechazando fechas fuera de ese margen. Los tests vencidos usan 60 segundos de antiguedad y los futuros inválidos 60 segundos de adelanto.

## Demo desplegada y documentación

- Frontend y API Go: https://api-go-matrix-challenge.onrender.com
- API Node: https://matrix-challenge-node.onrender.com
- Contratos HTTP: docs/API.md.
- Arquitectura del proyecto: docs/PROJECT_ARCHITECTURE.md.
- Decisiones funcionales y matemáticas: docs/ARCHITECTURE.md.
- Seguridad y tokens: docs/JWT.md.
- Evidencia de pruebas: docs/VALIDATION.md.
- Despliegue y reproducción: docs/DEPLOY_RENDER.md.
- Explicación paso a paso: docs/WALKTHROUGH.md.

Para probar la nube:

    $token = node --env-file=.env .\scripts\create-token.mjs
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1 -BaseUrl https://api-go-matrix-challenge.onrender.com -Token $token

Nota matemática: R es triangular superior cuando es cuadrada; para dimensiones rectangulares se describe como trapezoidal superior, con ceros en posiciones i>j.

Frontend: formulario simple con Bootstrap 5.3.8 y jQuery 4.0.0. Las opciones son Rotación y Rotación + factorización QR. Los recursos se incluyen en frontend/vendor; no se requiere un build frontend ni se cargan librerías de CDN en tiempo de ejecución.

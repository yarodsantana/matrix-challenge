# Contratos HTTP

Base Go/frontend: https://api-go-matrix-challenge.onrender.com
Base Node: https://matrix-challenge-node.onrender.com

En local, Go usa http://localhost:8080 y Node solo es accesible desde la red Compose. Las consultas usan Content-Type: application/json y Authorization: Bearer <token>. Go requiere token de cliente; Node requiere token de servicio. No son intercambiables.

## Go: POST /qr

Entrada: array de arrays de numeros finitos, rectangular y no vacío. Límite QR: 100 filas y 100 columnas. No se envuelve en {matrix:...}.

    [[2]]

Respuesta HTTP 200:

    {
      "original": [[2]],
      "rotated": [[2]],
      "q": [[-1]],
      "r": [[-2]],
      "statistics": {
        "max": -1,
        "min": -2,
        "average": -1.5,
        "sum": -3,
        "isDiagonal": true,
        "matrixCount": 2,
        "valueCount": 2,
        "diagonalMatrices": ["q", "r"],
        "diagonalTolerance": 1e-10,
        "perMatrix": {
          "q": {"max":-1,"min":-1,"average":-1,"sum":-1,"isDiagonal":true},
          "r": {"max":-2,"min":-2,"average":-2,"sum":-2,"isDiagonal":true}
        }
      },
      "metadata": {
        "factorization": "A = Q R",
        "qrMode": "full",
        "rotationDegrees": 90,
        "rotationDirection": "clockwise"
      }
    }

Para A de m x n, Q tiene m x m y R tiene m x n. QR se calcula sobre original, no sobre rotated. La rotación es una operación independiente. Las estadísticas principales consideran Q y R; no incluyen original ni rotated.

QR no tiene signos unicos. Se usa Householder sin forzar diagonal positiva en R; el producto y la ortogonalidad son las propiedades verificadas.

## Go: POST /rotate

Conserva el contrato anterior. Entrada:

    [[1,2,3],[4,5,6]]

Respuesta HTTP 200:

    {
      "matrix": [[4,1],[5,2],[6,3]],
      "statistics": {"max":6,"min":1,"average":3.5,"sum":21,"isDiagonal":false}
    }

Aqui matrix es la matriz rotada y las estadísticas se calculan sobre ella. La condicion diagonal conserva comparacion exacta con cero.

## Node: POST /statistics/qr

Go envia un objeto con ambas matrices:

    {"q":[[1,0],[0,1]],"r":[[2,3],[0,4]]}

Respuesta agregada:

    {
      "max":4,"min":0,"average":1.375,"sum":11,"isDiagonal":true,
      "matrixCount":2,"valueCount":8,"diagonalMatrices":["q"],
      "diagonalTolerance":1e-10,
      "perMatrix":{
        "q":{"max":1,"min":0,"average":0.5,"sum":2,"isDiagonal":true},
        "r":{"max":4,"min":0,"average":2.25,"sum":9,"isDiagonal":false}
      }
    }

Node valida matrices numéricas y dimensiones compatibles: Q cuadrada y R con la misma cantidad de filas. La generacion de la factorización corresponde a Go.

El promedio usa cantidad TOTAL de elementos. isDiagonal significa que alguna de las matrices es diagonal. Cada matriz debe ser cuadrada y cumplir abs(a[i][j]) <= 1e-10*max(1,maxAbsMatriz) cuando i!=j. Una matriz triangular superior puede no ser diagonal.

## Node: POST /statistics

Recibe una sola matriz directamente como array de arrays. Devuelve max, min, average, sum e isDiagonal. Lo usa Go para el modo /rotate. Todos los valores se incluyen, incluidos ceros.

## Salud y frontend

- GET /health en ambas APIs: HTTP 200 con {"status":"ok"}, sin JWT.
- GET / en Go: frontend HTML.
- GET /app.js y GET /styles.css en Go: assets del mismo origen.

/health comprueba que cada proceso responde; no reemplaza la prueba del recorrido entre servicios.

## Errores

Las respuestas de error de las APIs usan {"error":"mensaje"}. Los mensajes pueden variar; el cliente debe usar el código HTTP.

| Código | Significado |
| --- | --- |
| 400 | JSON o matriz inválida; dimensiones QR fuera del límite |
| 401 | Token ausente, inválido o vencido fuera de tolerancia |
| 404 | Ruta no encontrada |
| 413 | Cuerpo superior a 1 MiB |
| 415 | Content-Type no admitido |
| 422 | Resultado o suma fuera del rango numérico admitido |
| 502 | Node no disponible, estado inesperado o contrato remoto inválido |
| 504 | Timeout esperando a Node |

Los proxies de nube pueden generar sus propios errores fuera de la aplicación. Go no devuelve estadísticas inventadas ni exito parcial cuando Node falla. El cliente HTTP no sigue redirecciones ni reintenta automaticamente.

Nota matemática: R es triangular superior cuando es cuadrada; para dimensiones rectangulares se describe como trapezoidal superior, con ceros en posiciones i>j.

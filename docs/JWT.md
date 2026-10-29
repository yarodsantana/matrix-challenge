# JWT: Fase 3, bloque final

## Carpetas modificadas

- api-go/internal/auth/: configuracion, middleware y emision de tokens de servicio; pruebas.
- api-go/internal/statistics/: agrega Authorization al POST que Go envia a Express.
- api-go/cmd/api/: protege /rotate y adapta pruebas e integracion.
- api-node/src/auth.js: verifica los tokens de servicio antes de leer la matriz.
- api-node/src/app.js y api-node/test/: integra y prueba la autenticacion.
- scripts/: create-token.mjs, init-env.ps1 y smoke-test.ps1.
- raiz: compose.yaml, .gitignore, .env.example, README.md y PROMPTS.md.
- manifiestos y locks de Go y Node: dependencias JWT.

## Flujo de confianza

Cliente --JWT cliente--> Go --JWT servicio--> Node.

No se reenvia el JWT del cliente a Node. Go valida primero su firma y claims. Si la entrada es valida, rota la matriz y emite un token de servicio de dos minutos para Node. Node verifica el token antes de procesar el JSON.

| Campo | Cliente a Go | Go a Node |
| --- | --- | --- |
| algoritmo | HS256 | HS256 |
| clave | JWT_SECRET | SERVICE_JWT_SECRET |
| iss (emisor) | matrix-challenge | matrix-api-go |
| aud (destinatario) | matrix-api-go | matrix-api-node |
| sub (sujeto) | technical-test-user en la herramienta local | matrix-api-go |
| exp (vencimiento) | 15 minutos en la herramienta local | 2 minutos |

Go exige un sujeto no vacio; Node exige matrix-api-go. Ambas APIs exigen exp y validan su vencimiento, audiencia, emisor y algoritmo. Tambien rechazan nbf futuro e iat futuro cuando esos campos estan presentes. Sin token o con token invalido, devuelven 401 JSON y WWW-Authenticate: Bearer. Los mensajes no revelan detalles de la verificacion.

JWT firma los claims; no los cifra. El cuerpo de la matriz sigue viajando como JSON normal. Los poseedores de cada clave HMAC pueden firmar tokens para esa relacion de confianza. Las dos claves distintas evitan que la clave compartida con Node sea tambien la del acceso de clientes a Go.

GET /health queda sin autenticacion para los healthchecks. La ruta /rotate requiere JWT incluso cuando el JSON enviado es incorrecto: primero se autentica y despues se valida la entrada.

## Configuracion

Las claves deben tener al menos 32 bytes y ser distintas. No hay claves por defecto. Go se niega a iniciar con claves ausentes, cortas o iguales; Node rechaza una clave de servicio ausente o corta.

El .env local fue generado con dos claves aleatorias de 32 bytes representadas como 64 caracteres hexadecimales. No se muestran en las respuestas ni se incluyen en Git o en el contexto Docker. No copiar las claves publicas de las pruebas a .env.

Para una copia nueva del proyecto, desde la raiz:

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\init-env.ps1

El script no reemplaza un .env existente. .env.example documenta las variables; no contiene claves utilizables.

Compose lee .env y pasa JWT_SECRET y SERVICE_JWT_SECRET a Go. Node recibe solamente SERVICE_JWT_SECRET. Las claves se inyectan al ejecutar los contenedores; no se incrustan en las imagenes.

## Ejecutar con Docker

Desde la raiz del proyecto:

    docker compose up --build --wait --wait-timeout 120
    docker compose ps
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1

Es necesario reconstruir y recrear los contenedores anteriores: sus imagenes aun no incluyen JWT. El script ahora exige un 401 sin token y despues prueba los calculos con token. Necesita Node y las dependencias locales de api-node para generar el token; si faltan, ejecutar npm.cmd --prefix api-node ci.

## Token local y solicitud manual

Desde la raiz, con Node instalado y dependencias de api-node disponibles:

    $token = node --env-file=.env .\scripts\create-token.mjs
    Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/rotate' -Headers @{ Authorization = "Bearer $token" } -ContentType 'application/json' -Body '[[1,2,3],[4,5,6]]' | ConvertTo-Json -Depth 10

Para pasar un token ya emitido al smoke test:

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1 -Token $token

create-token.mjs es una herramienta local del desarrollador con acceso a JWT_SECRET. No existe un endpoint publico que entregue tokens sin verificar identidad. No se ha agregado una base de usuarios, login o proveedor externo: el requisito implementado es la autenticacion JWT de las dos APIs.

## Sintaxis, paso a paso

Go:
1. auth.New lee la configuracion validada una vez al crear la aplicacion.
2. security.Middleware es una funcion que Fiber ejecuta antes de rotateHandler.
3. jwt.ParseWithClaims verifica firma y claims con opciones explicitas; no basta con decodificar un token.
4. c.Next() permite continuar solo despues de autenticar.
5. client.Token es una funcion almacenada en una estructura; el cliente HTTP la invoca para obtener un token nuevo.
6. jwt.NewWithClaims(...).SignedString(...) firma el token que Node comprobara.

Node:
1. requireServiceToken(secret) devuelve una funcion middleware que recuerda la clave mediante un cierre (closure).
2. jwt.verify valida el token con algorithms, issuer, audience y subject explicitos.
3. exp se exige de forma adicional: la biblioteca verifica su valor cuando existe, pero nuestro contrato prohibe omitirlo.
4. next() continua al parser JSON y despues al calculo. Si falla la autenticacion, se responde 401 y se termina la solicitud.

## Pruebas

Node: npm.cmd --prefix api-node test.
Go, desde api-go: go test -tags integration ./... -count=1 y go vet ./....

Las pruebas usan claves publicas de prueba, independientes de .env. Se comprueban firma incorrecta, algoritmo distinto, expiracion, falta de exp, emisor/audiencia incorrectos, sujeto y fechas futuras. Las pruebas de calculo y errores de HTTP conservan su contrato usando tokens validos.

La integracion emite un token desde la herramienta Node, lo valida en Go y comprueba que Express acepta el token de servicio firmado por Go. En Windows, tambien ejecuta smoke-test.ps1 contra ambos servidores locales por HTTP real en puertos temporales.

Las imagenes Docker con JWT requieren validacion tras la reconstruccion desde una terminal con acceso al motor; la validacion local y la de configuracion se registran en PROMPTS.md.
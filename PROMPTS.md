P001 — Contrato y transformación verificable
“Antes de implementar HTTP, identifica ambigüedades del enunciado. Implementa en Go una rotación de 90° horario para matrices rectangulares, sin modificar la entrada. Rechaza matrices vacías o irregulares y demuestra el comportamiento con pruebas.”

Decisión humana: rotación horaria provisional, pendiente de confirmar con el enunciado.
Validación: comparar una matriz de 2 × 3 con una salida esperada de 3 × 2 y comprobar el rechazo de filas irregulares.

------------------

P002 — Adaptar la lógica a HTTP
“Expón la función de rotación existente mediante Fiber v3. Recibe un array de arrays directamente como JSON, valida el tipo de contenido y rechaza elementos null sin convertirlos en cero. Mantén la transformación fuera del manejador HTTP y define ejemplos de respuestas exitosas y errores.”

Decisión: usar POST /rotate y errores de entrada con estado 400; tipo de contenido no admitido con estado 415.
Evidencia pendiente: ejecución de las pruebas unitarias y comprobación de los ejemplos HTTP.

------------------

P003 — Verificar el contrato HTTP
“Escribe pruebas HTTP para la API de Go usando la configuración real de Fiber, sin abrir un puerto. Comprueba la rotación rectangular y el rechazo de matrices vacías, irregulares, elementos null, tipos incorrectos y JSON mal formado. Verifica códigos HTTP y cuerpos JSON.”

Decisión: separar la construcción de la aplicación de su ejecución mediante newApp.
Validación pendiente: ejecutar go test ./... -v y registrar el resultado real.
------------------

P004 — Completar la verificacion de la transformacion en el proyecto real
"Inspecciona el codigo existente antes de editar. Amplia las pruebas de Go para negativos, decimales, una fila, una columna, un elemento y entradas vacias. Comprueba que la rotacion conserva la entrada y que modificar la salida no cambia la matriz original. Ejecuta las pruebas y registra resultados reales."

Implementacion: nuevos casos en api-go/internal/matrix/rotate_test.go, sin cambiar el contrato HTTP ni la rotacion provisional de 90 grados horarios.
Verificacion: la suite existente paso; go vet ./... termino sin observaciones. La suite ampliada ejecuto sus aserciones correctamente con -cover, pero el informe fallo por permisos al renombrar archivos temporales del sandbox. Se reintenta con temporales dentro de api-go/work/go-tmp; resultado pendiente de registrar.

Resultado final P004: go test ./... -count=1 -cover paso en ambos paquetes usando temporales locales. Cobertura: cmd/api 90.9%; internal/matrix 100.0%. go vet ./... paso sin observaciones. Las pruebas HTTP verifican el endpoint de Go; aun no existe integracion con Node.js.

------------------

P005 — Estadisticas independientes de HTTP (Fase 2, bloque 1)
Estrategia aplicada: implementar una funcion pura de Node.js que valide matrices rectangulares numericas, calcule maximo, minimo, suma y promedio sobre todos los elementos y verifique la condicion diagonal. Probar resultados conocidos, negativos, decimales, matrices nulas, entradas invalidas y conservacion de la entrada antes de introducir Express.

Decision explicita: diagonal significa matriz cuadrada con ceros exactos fuera de la diagonal principal; una matriz cuadrada nula y una matriz 1x1 son diagonales. Esta convencion debe contrastarse con cualquier ejemplo adicional del enunciado.
Precision: Number y float64 usan doble precision; no se agrega redondeo. Se rechazan elementos no finitos y sumas intermedias no finitas.
Implementacion: api-node/src/statistics.js, api-node/test/statistics.test.js, api-node/package.json y api-node/README.md.
Verificacion real: node --test api-node/test/statistics.test.js con Node v24.22.1; 24 pruebas aprobadas y 0 fallidas. Aun no hay endpoint Express ni integracion Go-Node.

------------------

P006 — Exponer las estadisticas mediante Express y comprobar el contrato
Estrategia aplicada: separar createApp del arranque del servidor; recibir la matriz directamente como array de arrays JSON; reutilizar calculateStatistics sin duplicar calculos; devolver errores JSON con estados explicitos y probar solicitudes HTTP reales.

Archivos nuevos: api-node/src/app.js, api-node/src/server.js, api-node/test/app.test.js, api-node/.gitignore y api-node/package-lock.json.
Archivos actualizados: api-node/package.json (Express 5 con version exacta y script start), api-node/README.md y este registro.
Decision: POST /statistics, limite de 1 MiB, puerto 3000 configurable con PORT. Entrada invalida 400; cuerpo excesivo 413; contenido no admitido 415; desbordamiento de suma 422. Los fallos inesperados devuelven 500 sin exponer detalles al cliente.
Verificacion real: npm.cmd test aprobado. 24 casos unitarios y 16 escenarios HTTP; Node reporta 41 pruebas contando el contenedor de subpruebas, 0 fallidas. Los servidores de pruebas se cierran al finalizar.
Estado: Fase 2 implementada; pendiente Fase 3 (Go a Node por HTTP, Docker y JWT). La rotacion horaria y la convencion de matriz diagonal siguen siendo decisiones explicitas pendientes de contrastar con ejemplos adicionales del enunciado.

------------------

P007 — Comunicar Go con Express mediante HTTP
Estrategia aplicada: incorporar un cliente HTTP reutilizable en Go, enviar exclusivamente la matriz rotada a POST /statistics y devolver matriz y estadisticas al usuario. Mantener los calculos en Node, limitar tiempos y tamano de respuesta, y probar fallos del servicio remoto sin duplicar sus calculos.

Archivos nuevos: api-go/internal/statistics/client.go, api-go/cmd/api/integration_test.go y api-go/README.md. Modificados: api-go/cmd/api/main.go, api-go/cmd/api/main_test.go y este registro.
Contrato: Go conserva POST /rotate y ahora responde {matrix, statistics}. STATISTICS_URL por entorno, localhost:3000/statistics por defecto. Timeout de 3 segundos; errores 502/504 y 422 para suma fuera de rango. No se siguen redirecciones.
Aprendizaje: json.Marshal, http.NewRequestWithContext, cliente reutilizable, defer, errores envueltos con %w, errors.Is y dependencias inyectables.
Verificacion: suite Node aprobada (41 pruebas reportadas). go vet ./... aprobado. Integracion con Express real aprobada para rectangular y casos donde la rotacion cambia la condicion diagonal. Las pruebas de errores cubren timeout, conexion rechazada, estado 500, desbordamiento, JSON invalido, campos ausentes/null, tipos incorrectos, respuesta excesiva y JSON adicional.
Nota de entorno: una corrida encontro el limite predeterminado de 1 segundo de app.Test bajo carga; se amplio el tiempo del arnes de pruebas, manteniendo el timeout productivo de 3 segundos y la prueba de timeout especifica.
Estado: comunicacion HTTP implementada. Docker y JWT quedan para los proximos bloques.

------------------

P008 — Contenerizar y coordinar las dos APIs
Estrategia aplicada: crear Dockerfiles por servicio con pruebas durante la construccion, separar etapas de construccion y ejecucion, usar usuarios sin root y excluir dependencias de Windows y temporales. Definir Compose con STATISTICS_URL usando el nombre DNS api-node y esperar su estado healthy antes de iniciar Go.

Archivos nuevos: compose.yaml, README.md, api-go/Dockerfile, api-go/.dockerignore, api-node/Dockerfile, api-node/.dockerignore y scripts/smoke-test.ps1.
Archivos modificados: api-go/cmd/api/main.go, api-go/cmd/api/main_test.go, api-node/src/app.js, api-node/test/app.test.js y este registro. Se agrega GET /health en ambas APIs con pruebas.
Decision: solo Go publica puerto en el equipo (127.0.0.1:8080 por defecto, configurable con GO_PORT); Node permanece en la red Compose. El script verifica la matriz rotada y todas las estadisticas, incluida la condicion diagonal tras transformar.
Verificacion real: docker compose config --quiet aprobado; servicios reconocidos api-node y api-go; sintaxis PowerShell del script aprobada; suite Node con 42 pruebas reportadas, todas aprobadas.
Limitacion real: docker version no pudo conectar con dockerDesktopLinuxEngine por permission denied. docker compose build intento construir y fallo al abrir C:\Users\YAROD SANTANA\.docker\buildx\.lock por Access is denied. No se construyeron ni ejecutaron contenedores, ni se ejecuto el smoke test contra Docker. Los tags base tampoco se verificaron mediante una descarga exitosa.
Estado: configuracion Docker implementada y Compose validado; validacion real de contenedores pendiente. JWT aun pendiente.

Verificacion final P008: go test -tags integration ./... -count=1 aprobado tras agregar /health; ambos paquetes con pruebas pasaron. El paquete internal/statistics se ejercita desde las pruebas HTTP de cmd/api. Docker sigue pendiente por el acceso al motor.

------------------

P009 — Resolver el bloqueo local del smoke test sin cambiar politicas permanentes
Incidencia: el usuario construyo las dos imagenes y obtuvo ambos contenedores healthy, pero Windows PowerShell bloqueo la ejecucion directa del archivo .ps1.
Solucion: ejecutar powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1. La excepcion solo afecta al proceso temporal.
Evidencia del usuario: docker compose up construyo ambas imagenes y docker compose ps mostro los dos servicios healthy, Go publicado en 127.0.0.1:8080.
Verificacion real del asistente: el comando anterior termino con codigo 0; OK: rectangular; OK: diagonal tras rotar; Recorrido cliente -> Go -> Node -> Go verificado.
Archivos modificados: README.md (comandos y estado de validacion) y PROMPTS.md. No se modifico el codigo de las APIs ni la politica permanente de Windows.
Estado: bloque Docker verificado con evidencia de construccion del usuario y prueba HTTP ejecutada por el asistente. Siguiente bloque: JWT.

------------------

P010 — JWT en cliente-Go y Go-Node con contratos distintos
Estrategia aplicada: autenticar antes de validar matrices; fijar HS256 y comprobar firma, exp obligatorio, emisor, audiencia y sujeto; separar las claves de acceso de cliente y de servicio; emitir tokens internos cortos por solicitud; conservar /health publico y adaptar las pruebas de calculo existentes sin desactivar la autenticacion.

Diseno: JWT_SECRET protege cliente->Go (iss matrix-challenge, aud matrix-api-go, sub no vacio). SERVICE_JWT_SECRET protege Go->Node (iss y sub matrix-api-go, aud matrix-api-node). Herramienta local del desarrollador emite tokens cliente de 15 minutos; Go emite token de servicio de un minuto. Node recibe solo la clave de servicio en Compose. No se agrega endpoint publico de emision de tokens ni base de usuarios.
Configuracion: .env local creado con dos claves aleatorias distintas, sin mostrarlas en registros; .gitignore excluye secretos y .env.example contiene valores vacios. Las APIs rechazan configuracion ausente o corta; Go tambien rechaza claves iguales. Las claves se inyectan en runtime, no en las imagenes.

Carpetas: api-go/internal/auth (nuevo); api-go/internal/statistics y api-go/cmd/api (modificados); api-node/src y api-node/test (auth nuevo, app y pruebas modificadas); scripts (create-token.mjs e init-env.ps1 nuevos, smoke-test.ps1 actualizado); docs/JWT.md nuevo. Actualizados los manifiestos y locks, compose.yaml, READMEs y este registro.

Verificacion real: npm.cmd test reporto 57 pruebas, todas aprobadas. go test -tags integration ./... -count=1 -v paso en cmd/api, internal/auth e internal/matrix. go vet ./... paso. El paquete internal/statistics se ejercita desde las pruebas de cmd/api. Compose y sintaxis PowerShell validados.
Pruebas JWT: token valido, ausente, mal formado, esquema incorrecto, firma incorrecta, vencido, sin exp, aud/iss/sub incorrectos, emision futura, nbf futuro y algoritmo no permitido. Se comprueba que solicitudes no autenticadas no lleguen a Node y que /health siga disponible.
Integracion entre lenguajes: create-token.mjs emitio un token aceptado por Go; Go firmo un token de servicio aceptado por Express. En Windows, smoke-test.ps1 se ejecuto contra servidores temporales Go y Express reales por HTTP y reporto OK: solicitud sin token rechazada con 401, OK: rectangular, OK: diagonal tras rotar. Los procesos temporales se cerraron al terminar.

Incidencia de entorno resuelta: el cache global Go no permitia nuevas escrituras. Se uso cache local y GOPATH dentro de api-go/work, excluido de Git y Docker; un modulo temporal en work evita incluir dependencias descargadas en ./.... La verificacion criptografica de modulos permanece activa.
Estado: JWT implementado y verificado localmente. Los contenedores que el usuario construyo anteriormente corresponden al bloque sin JWT. Queda reconstruir y recrear con docker compose up --build --wait --wait-timeout 120 y repetir el smoke test autenticado para validar las nuevas imagenes en Docker.

------------------

P011 — Corregir alcance con el enunciado completo sin romper la implementacion previa
Estrategia aplicada: contrastar requisitos, conservar /rotate y agregar /qr. Resolver explicitamente la mezcla de rotacion y factorizacion: QR de la original, rotacion separada, Node analiza Q y R. Usar QR completa y Householder con escalado, validacion estructural compartida y limite 100x100. Exigir propiedades algebraicas y compatibilidad, no comparar una unica convencion de signos.
Implementacion: api-go/internal/matrix/qr.go y pruebas; handlers HTTP compartidos; cliente Node con contrato QR; api-node calcula agregados y perMatrix con tolerancia documentada. Se preservan matriz/statistics de /rotate y el comportamiento exacto de diagonal legacy.
Frontend: frontend/index.html, styles.css y app.js; servido por Go mediante internal/web. Entrada JSON, ejemplos, JWT en memoria, matrices A/rotacion/Q/R, estadisticas, error de reconstruccion, copia/descarga JSON. El navegador consume Go; Node permanece en la direccion HTTP exigida por el desafio.
Arquitectura y nombres: docs/ARCHITECTURE.md explica responsabilidades, ambiguedad, QR completa, signos, nombres camelCase, agregacion ponderada, tolerancia y diferencia Docker local/Render Free. No se entrega clave interna al navegador.
Verificacion real inicial: 67 pruebas Node aprobadas; Go con etiqueta integration aprobado; 64 matrices aleatorias deterministas y 12 escenarios QR de propiedades. El smoke por HTTP real verifico A=QR, Q ortogonal, R triangular superior, estadisticas agregadas, JWT y frontend disponible. La revision visual automatizada de navegador no esta disponible y no se afirma haberla hecho.

P012 — Preparar despliegue real en Render Free y repositorio GitHub
El usuario indico cuenta Render gratuita y creo https://github.com/yarodsantana/matrix-challenge.git. Se inicializo Git local y se configuro origin. .env y caches se comprobaron ignorados. GitHub CLI/conector no estan disponibles; el intento de consultar el remoto fallo al ejecutar el gestor de credenciales dentro del sandbox. No se han enviado credenciales al chat ni se ha publicado .env.
Preparacion: render.yaml con dos Web Services Free, contexto Docker correcto por servicio y secretos por entorno; docs/DEPLOY_RENDER.md con procedimiento manual Node primero, Go despues y prueba remota. No se inventan URLs publicadas. Se incluye scripts/publish.ps1 para el envio desde la terminal del usuario.
Cambio justificado: token de servicio ampliado de 1 a 2 minutos para cubrir timeout Render de 90s sin expirar durante el arranque en frio. Timeout local sigue siendo 3s; configurable 1..120s.
Estado: funcionalidades requeridas y frontend implementados localmente. Publicacion GitHub y despliegue cloud requieren autenticacion del usuario y validacion de URLs reales; no se declaran completados.

P013 — Revision de buenas practicas y cierre de calidad
Se separo configuracion de arranque y manejadores, eliminando panic del camino productivo. Se validan URL HTTP(S), timeout y claves antes de abrir el puerto. Se uniforman errores Fiber en JSON, se validan errores de resolucion de directorios y se agrupan imports Go. CSS pasa de compacto a formateado para facilitar lectura. La agregacion Node reutiliza un objeto perMatrix con nombres claros.
Pruebas nuevas: configuracion invalida, errores 404/413 JSON y contrato QR remoto (campos ausentes, conteo, diagonales y tipos). El 413 se prueba por HTTP real porque app.Test devuelve un error de transporte en ese caso. El healthcheck Node ahora respeta PORT.
Se agrega docs/WALKTHROUGH.md para estudiar sintaxis y defender arquitectura, QR, estadisticas y uso de IA. El usuario confirmo que publico el primer commit 2954b57 mediante autenticacion GitHub desde su propia terminal. El despliegue Render sigue en curso y no se afirma que este terminado.

P014 — Corregir el contexto Docker observado en Render
Evidencia del usuario: Deploy failed con failed to calculate checksum /test: not found. La causa es un Dockerfile Node que esperaba contexto api-node mientras Render usaba la raiz.
Correccion: ambos Dockerfiles usan ahora contexto raiz. Node copia api-node/package*.json, api-node/src y api-node/test; Compose y render.yaml se sincronizan. En Render, Root Directory vacio y Dockerfile Path api-node/Dockerfile o api-go/Dockerfile. No se modifican calculos ni JWT.
Esta correccion elimina la necesidad de configurar un contexto especial por servicio y evita ambiguedad del monorepo. Se espera confirmar un nuevo despliegue real despues de publicar el commit.

Verificacion final P013/P014: suite Go completa con integration aprobada en cmd/api, internal/auth, internal/matrix e internal/statistics. go vet aprobado. Compilacion Linux amd64 con CGO_ENABLED=0 aprobada. 67 pruebas Node aprobadas. Vista previa nativa con .env local paso smoke completo por HTTP, incluido QR y frontend. La prueba de limite HTTP se estabilizo comprobando rechazo de Content-Length excesivo antes del upload; se evito una carrera de cierre TCP de Windows. Compose validado tras unificar contextos. Despliegue Render aun pendiente de confirmar.

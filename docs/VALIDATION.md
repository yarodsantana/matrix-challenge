# Evidencia de validación

Fecha: 29 de octubre de 2026. Zona del proyecto: America/Lima.

## Servicios remotos verificados

| Componente | URL | Resultado observado |
| --- | --- | --- |
| Go + frontend | https://api-go-matrix-challenge.onrender.com | /health 200; pagina inicial 200; smoke completo aprobado |
| Node | https://matrix-challenge-node.onrender.com | /health 200; sin JWT 401; estadísticas autenticadas correctas |

Primera versión remota validada: 3b6b868. El propietario confirmo ambos servicios Live y se realizaron solicitudes HTTP reales. No se confunde un Dockerfile preparado con un despliegue completado.

## Prueba remota de extremo a extremo

Comando ejecutado desde el proyecto con un token cliente generado por la herramienta local:

    powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-test.ps1 -BaseUrl https://api-go-matrix-challenge.onrender.com -Token $token

Salida observada, código de salida 0:

    OK: solicitud sin token rechazada con 401
    OK: rectangular
    OK: diagonal tras rotar
    OK: QR 2 x 3, A=QR, Q ortogonal y estadisticas sobre Q+R
    OK: QR 2 x 2, A=QR, Q ortogonal y estadisticas sobre Q+R
    OK: frontend disponible
    Recorrido cliente -> Go -> Node -> Go verificado.

Esta prueba valida ambos servicios en la nube: las estadísticas de Go proceden de la llamada HTTP autenticada a Node.

Adicionalmente, Node recibio q identidad 2x2 y r [[2,3],[0,4]]: suma 11, promedio 1.375, máximo 4, mínimo 0 y 8 valores. Una solicitud sin JWT fue rechazada con 401.

## Pruebas locales automatizadas

| Comprobacion | Resultado |
| --- | --- |
| npm.cmd test en api-node | 68 pruebas reportadas, 68 aprobadas, 0 fallidas |
| go test -tags integration ./... -count=1 | cmd/api, internal/auth, internal/matrix e internal/statistics aprobados |
| go vet ./... | Aprobado en la revision de calidad |
| Build Go Linux amd64 con CGO_ENABLED=0 | Aprobado en la revision de portabilidad |
| docker compose config --quiet | Aprobado con contexto raíz para ambos Dockerfiles |
| Sintaxis JavaScript frontend | node --check aprobado |
| Sintaxis scripts PowerShell | Parser de PowerShell sin errores |

Las pruebas Node incluyen subpruebas; el total reportado por Node cuenta también sus grupos. No se usa ese numero como porcentaje de cobertura.

QR se verifica mediante reconstrucción A aproximadamente QR, ortogonalidad Q^T Q aproximadamente I, estructura triangular de R y conservacion de la entrada. Se incluyen 64 matrices aleatorias deterministas y casos cuadrados, altos, anchos, nulos, dependientes, decimales, negativos y rangos extremos.

JWT se verifica con claves de prueba independientes de las reales. Se comprueban firma, algoritmo, emisor, audiencia, sujeto y fechas. La revision final agrega tolerancia de reloj de 5 segundos, probada localmente: adelanto de 3s aceptado y adelanto/vencimiento de 60s rechazado. Debe desplegarse el ultimo commit para aplicar esta revision en Render.

## Límites de la evidencia

El frontend fue servido correctamente por HTTP, sus assets se probaron y su JavaScript paso validación de sintaxis. No se ejecuto una prueba visual automatizada de navegador; la interaccion puede verificarse manualmente pegando un JWT y calculando los ejemplos.

El entorno del asistente no tiene acceso directo al motor Docker. La ejecución de contenedores locales fue confirmada por el propietario; el despliegue Docker en Render y las pruebas remotas proporcionan evidencia adicional independiente.

Render Free puede suspender servicios inactivos. Antes de una demostracion conviene abrir ambos /health y generar un token nuevo. Las revisiones futuras deben repetir la prueba remota; esta evidencia no garantiza disponibilidad permanente del proveedor.

## Revisión del frontend con Bootstrap y jQuery

Se verificó la nueva interfaz con jQuery 4.0.0 real en un DOM simulado: validación de JSON, envío AJAX a /rotate y /qr, cabecera Bearer, presentación de resultados, recuperación del botón tras HTTP 401 y tratamiento seguro de mensajes como texto. Esta comprobación no equivale a una revisión visual en navegador. Bootstrap 5.3.8 y jQuery se sirven localmente; los assets se incluyen en las pruebas HTTP de Go.

## Emisión de tokens desde el frontend

Se agregaron pruebas para emisión con contraseña correcta, JWT emitido válido durante 15 minutos, autorización posterior de Go, contraseña incorrecta, JSON inválido, tipo de contenido, cuerpo excesivo, configuración ausente y sexto intento fallido limitado con HTTP 429. El DOM simulado verificó el formulario jQuery, la carga automática del token, el borrado del campo contraseña y la ausencia de almacenamiento local. La nueva ruta debe desplegarse y configurarse con DEMO_PASSWORD para validar este flujo en Render.

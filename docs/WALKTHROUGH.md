# Guia breve para defender y entender la solucion

## Recorrido de una solicitud

1. El reclutador usa Generar token con la contraseña compartida. Go emite el JWT y el frontend lo conserva en memoria. Después, el frontend convierte el texto JSON en arrays y valida estructura básica para dar feedback inmediato.
2. Envia POST /qr con Authorization: Bearer. La validación real y la autenticación siguen en el servidor.
3. Fiber verifica el JWT antes del manejador. readMatrix rechaza null, textos, matrices vacias o irregulares.
4. FactorQR calcula Q y R de la original. RotateClockwise calcula la rotación por separado.
5. El cliente HTTP Go emite un JWT de servicio y envia {q,r} a Express.
6. Express autentica, valida Content-Type y decodifica JSON. calculateQRStatistics calcula resultados.
7. Go comprueba el contrato recibido y devuelve matrices, estadísticas y metadatos.
8. El frontend muestra las matrices y estadísticas recibidas. jQuery gestiona el formulario y la petición AJAX; Bootstrap aporta los estilos. No se recalculan estadísticas en el navegador.

## Go que debes poder explicar

- [][]float64 es un slice de slices: no garantiza por si solo que todas las filas tengan la misma longitud.
- [][]*float64 se usa al leer JSON para distinguir null de cero. &body permite que el decodificador escriba en la variable; *value obtiene el numero después de comprobar nil.
- Las funciones devuelven resultado y error. Se comprueba err inmediatamente. Los errores de configuración se propagan al punto de entrada, que termina sin iniciar un servicio mal configurado.
- make reserva slices nuevos. Q, R y rotación no comparten memoria con la entrada.
- json.Marshal serializa; json.Unmarshal decodifica; la validación de campos obligatorios evita aceptar una respuesta {} como estadísticas de ceros.
- http.Client se reutiliza, tiene timeout y no sigue redirecciones. defer resp.Body.Close() libera la respuesta. %w conserva la causa para errors.Is.
- La funcion client.Token es una dependencia que genera tokens; las pruebas pueden aislar el transporte sin duplicar el código de autenticación.

## Householder sin memorizar toda la implementacion

Cada reflexion elimina los elementos que estan debajo de un pivote. Con un vector unitario v, H = I - 2 v v^T. Se aplica R <- HR y Q <- QH. Tras todas las reflexiones, A=QR y Q es ortogonal.

No se construye H como matriz completa: se usan productos con v. El signo se elige para evitar restar numeros casi iguales. math.Hypot y el escalado global evitan calcular normas mediante cuadrados que podrian desbordarse.

La QR no es unica: si cambias el signo de una columna de Q y de la fila correspondiente de R, el producto no cambia. Por eso las pruebas comprueban propiedades algebraicas. Se documenta que no se fuerza diagonal positiva en R; las estadísticas corresponden a esta convencion concreta.

## Node que debes poder explicar

- type: module habilita import/export; export permite probar las funciones sin levantar Express.
- const impide reasignar el enlace; let permite acumular suma, máximo y mínimo.
- Number.isFinite no convierte texto ni null a numero.
- Los middlewares se ejecutan en orden: autenticar, comprobar contenido, parsear, calcular y manejar errores.
- next() continua; next(error) entrega el fallo al manejador de cuatro parametros.
- requireServiceToken devuelve una funcion que conserva la clave mediante un cierre (closure).
- El promedio global usa cantidad total de valores. No se promedian promedios sin ponderar.

## Preguntas previsibles

### Por que Go no calcula también las estadísticas?

Porque el desafio asigna esa responsabilidad a Node. Mantenerla separada permite probar cada servicio, cambiarlo sin duplicar calculos y demostrar comunicación HTTP real.

### Por que el frontend no llama directamente a Node?

La arquitectura exige que Go envie los resultados procesados. El navegador recibe la respuesta de ambos a traves de Go y no necesita la credencial interna ni CORS.

### Por que agregaste /qr en lugar de modificar /rotate?

Para cubrir el requisito QR que aparecio en el enunciado completo sin romper clientes ni pruebas del contrato que ya existia. El frontend ofrece ambos modos y distingue los resultados.

### Una matriz triangular superior es diagonal?

No. Diagonal exige que sea cuadrada y que TODOS los elementos fuera de la diagonal sean cero. Una triangular puede tener valores por encima. En QR se usa una tolerancia numérica documentada.

### Que ocurre si Node falla?

Go devuelve 502 si no hay conexión o la respuesta es inválida, 504 ante timeout y 422 ante rechazo por desbordamiento numérico. No se inventan estadísticas ni se devuelve exito parcial.

### Por que dos secretos JWT?

Son dos relaciones de confianza diferentes. Un token cliente esta destinado a Go; un token de servicio, a Node. Se comprueban algoritmo, firma, emisor, audiencia y vencimiento. El token firma los claims, no los cifra.

### Por que frontend sin React?

La interfaz necesita un formulario, tablas y una llamada HTTP. HTML/CSS/JS reduce dependencias y compilacion, facilita explicar el código y permite servirlo desde Go. No cambia los frameworks exigidos a las APIs.

### Que limita esta implementacion?

QR hasta 100x100, float64/Number con precisión finita, convencion QR completa, arranque en frio de Render Free y acceso de demostración con contraseña compartida. No es un sistema de identidad con usuarios o renovacion de sesiones.

## Desarrollo asistido por IA

Mantener un registro personal de prompts y decisiones, fuera del repositorio y de Docker, no como sustituto de comprender el código. Las evidencias publicas de ejecución estan en VALIDATION.md. Explicar que la IA ayudo a dividir el problema, detectar ambiguedades, generar implementacion y proponer casos de prueba; las propiedades algebraicas, contratos y evidencias de ejecución permiten revisar sus propuestas. Distinguir siempre código preparado de despliegue efectivamente verificado.

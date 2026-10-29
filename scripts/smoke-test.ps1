param([string]$BaseUrl = 'http://localhost:8080', [string]$Token)
$ErrorActionPreference = 'Stop'
if ([string]::IsNullOrWhiteSpace($Token)) {
 $root = Split-Path -Parent $PSScriptRoot
 $Token = & node "--env-file=$(Join-Path $root '.env')" (Join-Path $PSScriptRoot 'create-token.mjs')
 if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($Token)) { throw 'No se pudo generar el token local' }
}
$headers = @{ Authorization = "Bearer $Token" }
# Verifica primero que la ruta ya no acepta solicitudes anonimas.
try {
 Invoke-WebRequest -UseBasicParsing -Method Post -Uri "$BaseUrl/rotate" -ContentType 'application/json' -Body '[[1]]' -TimeoutSec 120 | Out-Null
 throw 'La API acepto una solicitud sin JWT; reconstruye los contenedores'
} catch {
 if ($null -eq $_.Exception.Response -or [int]$_.Exception.Response.StatusCode -ne 401) { throw }
 Write-Output 'OK: solicitud sin token rechazada con 401'
}
$health = Invoke-RestMethod -Uri "$BaseUrl/health" -TimeoutSec 120
if ($health.status -ne 'ok') { throw 'Go no responde correctamente a /health' }

$cases = @(
 @{ Name = 'rectangular'; Input = '[[1,2,3],[4,5,6]]'; Matrix = '[[4,1],[5,2],[6,3]]'; Max = 6; Min = 1; Sum = 21; Average = 3.5; Diagonal = $false },
 @{ Name = 'diagonal tras rotar'; Input = '[[0,1],[2,0]]'; Matrix = '[[2,0],[0,1]]'; Max = 2; Min = 0; Sum = 3; Average = 0.75; Diagonal = $true }
)
foreach ($case in $cases) {
 $result = Invoke-RestMethod -Method Post -Uri "$BaseUrl/rotate" -Headers $headers -ContentType 'application/json' -Body $case.Input -TimeoutSec 120
 $actualMatrix = ConvertTo-Json -InputObject $result.matrix -Compress -Depth 10
 $s = $result.statistics
 if ($actualMatrix -ne $case.Matrix -or $s.max -ne $case.Max -or $s.min -ne $case.Min -or $s.sum -ne $case.Sum -or $s.average -ne $case.Average -or $s.isDiagonal -ne $case.Diagonal) {
  throw "Respuesta incorrecta para $($case.Name): $(ConvertTo-Json $result -Compress -Depth 10)"
 }
 Write-Output "OK: $($case.Name)"
}
# Flujo principal QR: reconstruccion, ortogonalidad y estadisticas de Q + R.
foreach ($inputJSON in @('[[1,2,3],[4,5,6]]','[[2,0],[0,3]]')) {
 $original = ConvertFrom-Json $inputJSON
 $qr = Invoke-RestMethod -Method Post -Uri "$BaseUrl/qr" -Headers $headers -ContentType 'application/json' -Body $inputJSON -TimeoutSec 120
 $rows = $original.Count
 $cols = $original[0].Count
 if ($qr.q.Count -ne $rows -or $qr.q[0].Count -ne $rows -or $qr.r.Count -ne $rows -or $qr.r[0].Count -ne $cols) { throw 'Dimensiones QR incorrectas' }
 for ($i=0; $i -lt $rows; $i++) {
  for ($j=0; $j -lt $cols; $j++) {
   $value=0.0
   for ($k=0; $k -lt $rows; $k++) { $value += $qr.q[$i][$k] * $qr.r[$k][$j] }
   if ([Math]::Abs($value-$original[$i][$j]) -gt 1e-8) { throw 'Q por R no reconstruye A' }
   if ($i -gt $j -and [Math]::Abs($qr.r[$i][$j]) -gt 1e-10) { throw 'R no es triangular superior' }
  }
  for ($j=0; $j -lt $rows; $j++) {
   $dot=0.0
   for ($k=0; $k -lt $rows; $k++) { $dot += $qr.q[$k][$i] * $qr.q[$k][$j] }
   $expected=0.0
   if ($i -eq $j) { $expected=1.0 }
   if ([Math]::Abs($dot-$expected) -gt 1e-8) { throw 'Q no es ortogonal' }
  }
 }
 $sum=0.0; $max=[double]::NegativeInfinity; $min=[double]::PositiveInfinity; $count=0
 foreach ($name in @('q','r')) {
  foreach ($row in $qr.$name) {
   foreach ($value in $row) { $sum+=$value; $max=[Math]::Max($max,$value); $min=[Math]::Min($min,$value); $count++ }
  }
 }
 $s=$qr.statistics
 if ([Math]::Abs($s.sum-$sum) -gt 1e-8 -or [Math]::Abs($s.average-$sum/$count) -gt 1e-8 -or [Math]::Abs($s.max-$max) -gt 1e-8 -or [Math]::Abs($s.min-$min) -gt 1e-8 -or $s.valueCount -ne $count) { throw 'Estadisticas QR incorrectas' }
 if ($s.isDiagonal -ne ($s.perMatrix.q.isDiagonal -or $s.perMatrix.r.isDiagonal)) { throw 'Indicador de alguna matriz diagonal incorrecto' }
 Write-Output "OK: QR $rows x $cols, A=QR, Q ortogonal y estadisticas sobre Q+R"
}
$page = Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/" -TimeoutSec 120
if ($page.StatusCode -ne 200 -or !$page.Content.Contains('Matrix')) { throw 'Frontend no disponible' }
Write-Output 'OK: frontend disponible'
Write-Output 'Recorrido cliente -> Go -> Node -> Go verificado.'
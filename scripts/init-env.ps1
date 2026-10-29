$ErrorActionPreference = 'Stop'
$path = Join-Path (Split-Path -Parent $PSScriptRoot) '.env'
if (Test-Path -LiteralPath $path) { Write-Output '.env ya existe; no se modifica.'; exit 0 }
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
try {
 $bytes = New-Object byte[] 32
 $rng.GetBytes($bytes)
 $clientSecret = [BitConverter]::ToString($bytes).Replace('-','').ToLowerInvariant()
 $rng.GetBytes($bytes)
 $serviceSecret = [BitConverter]::ToString($bytes).Replace('-','').ToLowerInvariant()
 $rng.GetBytes($bytes)
 $demoSecret = [BitConverter]::ToString($bytes).Replace('-','').ToLowerInvariant()
 [IO.File]::WriteAllText($path, "JWT_SECRET=$clientSecret`nSERVICE_JWT_SECRET=$serviceSecret`nDEMO_PASSWORD=$demoSecret`nGO_PORT=8080`n")
 Write-Output '.env creado con dos claves aleatorias; no compartir ni subir a Git.'
} finally { $rng.Dispose() }
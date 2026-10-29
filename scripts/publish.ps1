$ErrorActionPreference='Stop'
$root=Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root
$gitCommand=Get-Command git -ErrorAction SilentlyContinue
if ($gitCommand -and $gitCommand.Source) { $gitBinary=$gitCommand.Source } else {
 $gitBinary=Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\native\git\cmd\git.exe'
 if (!(Test-Path -LiteralPath $gitBinary)) { throw 'Instala Git o abre una terminal con git disponible.' }
}
& $gitBinary push -u origin main
if ($LASTEXITCODE -ne 0) { throw 'Git no pudo publicar; revisa tu autenticacion y permisos del repositorio.' }
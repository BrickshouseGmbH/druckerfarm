$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Build aus dem Wurzelordner (flache Struktur). Version aus VERSION, Build-ID =
# Kompilier-Zeitstempel (automatisch, immer eindeutig).
$version = (Get-Content VERSION -Raw).Trim()
$buildId = Get-Date -Format "yyyy-MM-dd-HHmmss"
Write-Host "Baue druckerfarm.exe ($version, Build $buildId)..." -ForegroundColor Cyan

go build -trimpath -mod=vendor -ldflags="-s -w -H windowsgui -X main.appVersion=$version -X main.buildID=$buildId" -o druckerfarm.exe .

if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item druckerfarm.exe).Length / 1MB, 1)
    Write-Host "Fertig! druckerfarm.exe $version ($size MB)" -ForegroundColor Green
} else {
    Write-Host "Build fehlgeschlagen!" -ForegroundColor Red
}

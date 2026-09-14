$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$version = (Get-Content VERSION -Raw).Trim()
Write-Host "Baue druckerfarm.exe ($version)..." -ForegroundColor Cyan
go build -trimpath -mod=vendor -ldflags="-s -w -H windowsgui -X main.appVersion=$version" -o druckerfarm.exe .
if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item druckerfarm.exe).Length / 1MB, 1)
    Write-Host "Fertig! druckerfarm.exe $version ($size MB)" -ForegroundColor Green
} else {
    Write-Host "Build fehlgeschlagen!" -ForegroundColor Red
}

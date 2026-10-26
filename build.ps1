$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Sauberer, absturzsicherer Release-Build.
# Der Go-Code liegt in app/, VERSION im Hauptverzeichnis, die fertige
# druckerfarm.exe ebenfalls hier. Läuft die App noch, ist druckerfarm.exe
# gesperrt — dann wird nach druckerfarm.new.exe gebaut und getauscht, sobald
# die App geschlossen ist. So gibt es keine verwaisten build0N-Varianten mehr.

$version = (Get-Content VERSION -Raw).Trim()
# Build-ID = Kompilier-Zeitstempel (YYYY-MM-DD-HHMMSS). Wird automatisch gesetzt,
# ist immer eindeutig und muss nie von Hand hochgezählt werden.
$buildId = Get-Date -Format "yyyy-MM-dd-HHmmss"
$tmp = "druckerfarm.new.exe"
Write-Host "Baue druckerfarm.exe ($version, Build $buildId)..." -ForegroundColor Cyan

Push-Location app
go build -trimpath -mod=vendor -ldflags="-s -w -H windowsgui -X main.appVersion=$version -X main.buildID=$buildId" -o "..\$tmp" .
$code = $LASTEXITCODE
Pop-Location

if ($code -ne 0) { Write-Host "Build fehlgeschlagen!" -ForegroundColor Red; exit 1 }

# Herstellername darf nicht im Binary auftauchen.
$hit = Select-String -Path $tmp -Pattern "bambu" -SimpleMatch -CaseSensitive:$false -List
if ($hit) { Write-Host "WARNUNG: 'bambu' im Binary gefunden!" -ForegroundColor Red }

# Tauschen, sobald möglich; sonst .new.exe stehen lassen.
$swapped = $false
try {
    if (Test-Path druckerfarm.exe) { Remove-Item druckerfarm.exe -ErrorAction Stop }
    Rename-Item $tmp druckerfarm.exe -ErrorAction Stop
    $swapped = $true
} catch {
    Write-Host "druckerfarm.exe ist gesperrt (App läuft?) — neue Fassung liegt als $tmp bereit." -ForegroundColor Yellow
    Write-Host "App schließen und dieses Skript erneut ausführen, dann wird getauscht." -ForegroundColor Yellow
}

# Verwaiste Varianten aufräumen (alte build0N-/neu-Dateien).
Get-ChildItem -File | Where-Object { $_.Name -match '^druckerfarm[.\-].*\.exe$' -and $_.Name -ne 'druckerfarm.new.exe' } | Remove-Item -ErrorAction SilentlyContinue

$target = if ($swapped) { "druckerfarm.exe" } else { $tmp }
$size = [math]::Round((Get-Item $target).Length / 1MB, 1)
Write-Host "Fertig! $target $version ($size MB)" -ForegroundColor Green

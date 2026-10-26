#!/bin/bash
# Der Go-Code liegt im Unterordner app/. Version kommt aus der Datei VERSION
# im Hauptverzeichnis; die Exe landet ebenfalls im Hauptverzeichnis.
set -e
VERSION=$(tr -d '[:space:]' < VERSION)
cd "$(dirname "$0")/app"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -trimpath -mod=vendor -ldflags="-s -w -H windowsgui -X main.appVersion=$VERSION" \
  -o ../druckerfarm.exe .
echo "Done: druckerfarm.exe $VERSION"

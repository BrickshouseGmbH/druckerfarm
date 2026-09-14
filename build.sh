#!/bin/bash
VERSION=$(tr -d '[:space:]' < VERSION)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -trimpath -mod=vendor -ldflags="-s -w -H windowsgui -X main.appVersion=$VERSION" \
  -o druckerfarm.exe .
echo "Done: druckerfarm.exe $VERSION"

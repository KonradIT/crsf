@echo off
setlocal

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -tags usejoystick -ldflags="-s -w" -trimpath -o joystick.exe .\examples\joystick\main.go
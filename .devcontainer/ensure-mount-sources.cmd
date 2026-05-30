@echo off

REM This Windows batch script ensures that the default source mount points in
REM devcontainer.json exist on the host.

setlocal enableextensions
echo "Ensuring mount points exist..."
md "%USERPROFILE%\.kube"
endlocal

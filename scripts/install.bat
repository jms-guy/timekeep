@echo off
echo TimeKeep Installer
echo ==================
echo This will install TimeKeep as a Windows service.
echo You may need to approve Administrator privileges.
echo.
pause

PowerShell -ExecutionPolicy Bypass -File "%~dp0install.ps1"
pause
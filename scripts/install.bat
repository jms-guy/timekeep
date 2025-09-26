@echo off
echo TimeKeep Installer
echo ==================
echo.
echo This will open PowerShell as Administrator.
echo Please approve the UAC prompt.
echo.
pause

PowerShell -Command "Start-Process PowerShell -ArgumentList '-NoExit -ExecutionPolicy Bypass -File \"%~dp0install.ps1\"' -Verb RunAs"
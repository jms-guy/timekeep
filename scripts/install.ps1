#Requires -RunAsAdministrator

if ($PSScriptRoot) {
    Set-Location $PSScriptRoot
}

if (Test-Path "windows\timekeep-service.exe") {
    $BinaryDir = "windows"
} elseif (Test-Path "timekeep-service.exe") {
    $BinaryDir = "."
} else {
    Write-Host "Error: Timekeep binaries not found."
    Write-Host "Please extract the release ZIP and run this script from the extracted directory"
    exit 1
}

Write-Host "Installing Timekeep..."

$InstallPath = "C:\Program Files\TimeKeep"
New-Item -ItemType Directory -Force -Path $InstallPath

Copy-Item "$BinaryDir\timekeep-service.exe" -Destination "$InstallPath\"
Copy-Item "$BinaryDir\timekeep.exe" -Destination "$InstallPath\"

$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$InstallPath*") {
    $NewPath = $CurrentPath + ";" + $InstallPath
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
}

$env:Path = $env:Path + ";" + $InstallPath

sc.exe create TimeKeep binPath= "$InstallPath\timekeep-service.exe" start= auto
sc.exe start TimeKeep

Write-Host "Installation complete! Run 'timekeep ping' to test."
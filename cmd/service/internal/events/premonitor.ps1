<#
    This script runs before the main process monitoring script. It queries for processes that belong to programs being tracked
    that are already running, and sends the service a synthetic "process_start" event, immediately opening an active session for
    that program.
#>
param(
    [string]$Programs
)

# Fail fast on errors
$ErrorActionPreference = "Stop"

# Connect to named pipe opened by service
$pipe = New-Object System.IO.Pipes.NamedPipeClientStream(".", "Timekeep", "Out")
$pipe.Connect()
$writer = New-Object System.IO.StreamWriter($pipe)
$writer.AutoFlush = $true

try {
    $tracked = $Programs -split ","

    if (-not $tracked -or $tracked.Count -eq 0) {
        exit 0
    }

    # Build a set for quick membership checks
    $set = @{}
    foreach ($n in $tracked) { $set[$n] = $true }

    # Enumerate current processes and emit synthetic start events
    Get-CimInstance Win32_Process | ForEach-Object {
        $name = $_.Name
        if ($name -and $set.ContainsKey($name.ToLower())) {
            $data = @{
                action = "process_start"
                name   = $name
                pid    = [int]$_.ProcessId
            }
            $writer.WriteLine(($data | ConvertTo-Json -Compress))
        }
    }
}
catch {
    # Surface errors to logs
    $err = @{
        action = "ps_error"
        message = $_.Exception.Message
    }
    $writer.WriteLine(($err | ConvertTo-Json -Compress))
    exit 1
}
finally {
    $writer.Flush()
    $writer.Dispose()
    $pipe.Dispose()
}
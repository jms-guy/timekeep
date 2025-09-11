<#
    This script will query program names issued as command line arguments, and register for WMI start/stop events
    related to those specific programs. Writing any actions returned to a named pipe opened by the service
#>

param(
    [string]$Programs
)

# Connect to named pipe opened by service
$pipe = New-Object System.IO.Pipes.NamedPipeClientStream(".", "Timekeep", "Out")
$pipe.Connect()
$writer = New-Object System.IO.StreamWriter($pipe)

# Get programs to track from arguments
$trackedPrograms = $Programs -split ","
$whereClause = ($trackedPrograms | ForEach-Object { "ProcessName='$_'" }) -join " OR "

$startQuery = "SELECT * FROM Win32_ProcessStartTrace WHERE $whereClause"
$stopQuery = "SELECT * FROM Win32_ProcessStopTrace WHERE $whereClause"

# Register for WMI events
Register-WmiEvent -Query $startQuery -Action {
    $processName = $Event.SourceEventArgs.NewEvent.ProcessName
    $processID = $Event.SourceEventArgs.NewEvent.ProcessID
    
    $data = @{
        action = "process_start"
        name = $processName
        pid = $processID
    }
    $writer.WriteLine(($data | ConvertTo-Json -Compress))
    $writer.Flush()
}

Register-WmiEvent -Query $stopQuery -Action {
    $processName = $Event.SourceEventArgs.NewEvent.ProcessName
    $processID = $Event.SourceEventArgs.NewEvent.ProcessID
    
    $data = @{
        action = "process_stop"
        name = $processName
        pid = $processID
    }
    $writer.WriteLine(($data | ConvertTo-Json -Compress))
     $writer.Flush()
}

while ($true) {
    Start-Sleep -Seconds 1
}
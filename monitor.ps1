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

# Build queries
$startQuery = "SELECT * FROM Win32_ProcessStartTrace WHERE $whereClause"
$stopQuery = "SELECT * FROM Win32_ProcessStopTrace WHERE $whereClause"

# Register for WMI start events
Register-WmiEvent -Query $startQuery -Action {
    $data = @{action="process_start"; name=$Event.SourceEventArgs.NewEvent.ProcessName}
    $writer.WriteLine(($data | ConvertTo-Json))
}

# Register for WMI stop events
Register-WmiEvent -Query $stopQuery -Action {
    $data = @{action="process_stop"; name=$Event.SourceEventArgs.NewEvent.ProcessName}
    $writer.WriteLine(($data | ConvertTo-Json))
}
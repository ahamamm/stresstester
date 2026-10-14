# Define the source folder path and the parent destination directory
$SourceFolder = "D:\code\GoCode\stresstester"
$DestinationParent = "D:\code_bk\GoCode_bk\stresstester_bk"

# Generate the current date and time string in YYYYMMDDHHmm format
$Timestamp = Get-Date -Format "yyyyMMddHHmm"

# Combine the parent destination path with the new timestamp folder name
$DestinationFolder = Join-Path -Path $DestinationParent -ChildPath $Timestamp

# Copy the folder and all its contents to the new location
Copy-Item -Path $SourceFolder -Destination $DestinationFolder -Recurse -Force

# Make the new folder and all files inside it Read-Only
Get-ChildItem -Path $DestinationFolder -Recurse -File | ForEach-Object { $_.IsReadOnly = $true }

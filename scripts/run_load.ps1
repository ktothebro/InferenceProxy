param(
  [string]$Url = "http://localhost:8080/infer",
  [string]$Dur = "20s",
  [int]$C = 200
)
Write-Host "Load: $Url  duration=$Dur  concurrency=$C"
if (Test-Path ".\hey.exe") {
  .\hey.exe -z $Dur -c $C -m POST -d '{"q":"test"}' $Url
} else {
  Write-Host "Place hey.exe in the repo root, or update this script to point to it."
}

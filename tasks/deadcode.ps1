# Check for unreachable functions in cmd entry points
$out = deadcode ./examples/... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host $out
    exit 1
}
if ($out) {
    Write-Host "dead code found:"
    Write-Host $out
    exit 1
}
Write-Host "no dead code found"

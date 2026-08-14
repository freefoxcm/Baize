param()

$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))

Push-Location $repoRoot
try {
    go test `
        ./internal/cli `
        ./internal/serve `
        ./internal/i18n `
        ./internal/boot `
        ./internal/plugin `
        -count=1
    if ($LASTEXITCODE -ne 0) {
        throw "Focused Windows tests failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

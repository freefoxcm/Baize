$ErrorActionPreference = "Stop"

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

$repoRoot = (& git rev-parse --show-toplevel 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $repoRoot) {
    throw "run this script from inside the Reasonix repository"
}
Set-Location $repoRoot

Invoke-Git remote get-url upstream | Out-Null
$gitDir = (& git rev-parse --git-dir).Trim()
$hookDir = Join-Path $gitDir "baize-hooks"
New-Item -ItemType Directory -Force -Path $hookDir | Out-Null
Copy-Item -Force .githooks/pre-commit, .githooks/pre-push -Destination $hookDir
$hookConfigPath = $hookDir.Replace('\', '/')

Invoke-Git config --local rerere.enabled true
Invoke-Git config --local rerere.autoupdate true
Invoke-Git config --local pull.ff only
Invoke-Git config --local fetch.prune true
Invoke-Git config --local core.hooksPath $hookConfigPath
Invoke-Git config --local merge.baize.name "Keep Baize-owned frontend during controlled upstream merges"
Invoke-Git config --local merge.baize.driver true
Invoke-Git remote set-url --push upstream DISABLED

Write-Host "Baize fork Git settings installed for $repoRoot"
Write-Host "upstream is fetch-only; main-v2 is the mirror and custom/baize is the integration branch"

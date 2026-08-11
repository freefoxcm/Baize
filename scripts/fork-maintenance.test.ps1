$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$testRoot = Join-Path ([IO.Path]::GetTempPath()) ("baize-fork-test-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $testRoot | Out-Null

function Invoke-TestGit {
    param([string]$Directory, [Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & git.exe -C $Directory @Arguments
    if ($LASTEXITCODE -ne 0) { throw "git failed: $($Arguments -join ' ')" }
}

try {
    $upstream = Join-Path $testRoot "upstream.git"
    $origin = Join-Path $testRoot "origin.git"
    $seed = Join-Path $testRoot "seed"
    $work = Join-Path $testRoot "work"
    & git.exe init -q --bare $upstream
    & git.exe init -q --bare $origin
    & git.exe init -q -b main-v2 $seed
    Invoke-TestGit $seed config user.name test
    Invoke-TestGit $seed config user.email test@example.com
    Set-Content -LiteralPath (Join-Path $seed "value.txt") -Value "one" -NoNewline
    Invoke-TestGit $seed add value.txt
    Invoke-TestGit $seed commit -qm one
    Invoke-TestGit $seed remote add upstream $upstream
    Invoke-TestGit $seed remote add origin $origin
    Invoke-TestGit $seed push -q upstream main-v2
    Invoke-TestGit $seed push -q origin main-v2
    & git.exe clone -q -b main-v2 $origin $work
    Invoke-TestGit $work config user.name test
    Invoke-TestGit $work config user.email test@example.com
    Invoke-TestGit $work remote add upstream $upstream
    Invoke-TestGit $work switch -qc custom/baize
    New-Item -ItemType Directory -Path (Join-Path $work "scripts") | Out-Null
    Copy-Item (Join-Path $repoRoot "scripts/sync-upstream.ps1"), (Join-Path $repoRoot "scripts/baize-workflows-active.txt"), (Join-Path $repoRoot "scripts/baize-workflows-disabled.txt") -Destination (Join-Path $work "scripts")
    Invoke-TestGit $work add scripts
    Invoke-TestGit $work commit -qm scripts
    $customBefore = (& git.exe -C $work rev-parse custom/baize).Trim()
    $originBefore = (& git.exe --git-dir=$origin rev-parse main-v2).Trim()
    Add-Content -LiteralPath (Join-Path $seed "value.txt") -Value "two"
    Invoke-TestGit $seed commit -qam two
    Invoke-TestGit $seed push -q upstream main-v2
    $upstreamTip = (& git.exe -C $seed rev-parse HEAD).Trim()
    Push-Location $work
    try {
        & pwsh -NoProfile -File (Join-Path $work "scripts/sync-upstream.ps1") | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "PowerShell sync failed" }
    }
    finally { Pop-Location }
    if ((& git.exe -C $work branch --show-current).Trim() -ne "custom/baize") { throw "sync did not restore the branch" }
    if ((& git.exe -C $work rev-parse main-v2).Trim() -ne $upstreamTip) { throw "main-v2 was not updated" }
    if ((& git.exe -C $work rev-parse custom/baize).Trim() -ne $customBefore) { throw "custom/baize changed" }
    if ((& git.exe --git-dir=$origin rev-parse main-v2).Trim() -ne $originBefore) { throw "sync pushed origin" }

    $github = Join-Path $testRoot "github"
    New-Item -ItemType Directory -Path (Join-Path $github "scripts") -Force | Out-Null
    Copy-Item (Join-Path $repoRoot "scripts/setup-fork-github.ps1"), (Join-Path $repoRoot "scripts/baize-workflows-active.txt"), (Join-Path $repoRoot "scripts/baize-workflows-disabled.txt") -Destination (Join-Path $github "scripts")
    & git.exe init -q -b custom/baize $github
    Invoke-TestGit $github remote add origin https://github.com/freefoxcm/DeepSeek-Reasonix.git
    $fakeGh = Join-Path $testRoot "fake-gh.ps1"
    @'
$joined = $args -join ' '
if ($joined -eq 'auth status') { exit 0 }
if ($joined -like 'repo view *') { 'custom/baize'; exit 0 }
if ($joined -like 'workflow list *') {
  '[{"path":".github/workflows/baize-cache-impact.yml","state":"active"},{"path":".github/workflows/baize-ci.yml","state":"active"},{"path":".github/workflows/baize-docs-impact.yml","state":"active"}]'
  exit 0
}
if ($joined -like 'workflow disable *' -or $joined -like 'workflow enable *') { exit 0 }
if ($joined -like 'api repos/freefoxcm/DeepSeek-Reasonix/actions/permissions*') { 'true'; exit 0 }
if ($joined -like 'api repos/freefoxcm/DeepSeek-Reasonix --jq*') { 'true'; exit 0 }
if ($args[0] -eq 'api') { '{}'; exit 0 }
throw "unexpected fake gh invocation: $joined"
'@ | Set-Content -LiteralPath $fakeGh
    $env:BAIZE_GH_BIN = $fakeGh
    Push-Location $github
    try {
        & pwsh -NoProfile -File (Join-Path $github "scripts/setup-fork-github.ps1") -Check | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "GitHub check-mode test failed" }
        & pwsh -NoProfile -File (Join-Path $github "scripts/setup-fork-github.ps1") -Apply | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "GitHub apply-mode test failed" }
    }
    finally { Pop-Location }
    Remove-Item Env:BAIZE_GH_BIN

    Write-Output "fork maintenance PowerShell tests passed"
}
finally {
    if (Test-Path -LiteralPath $testRoot) { Remove-Item -LiteralPath $testRoot -Recurse -Force }
}

$ErrorActionPreference = "Stop"

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

function Test-GitPath([string]$Name) {
    $path = (& git rev-parse --git-path $Name 2>$null)
    return $LASTEXITCODE -eq 0 -and (Test-Path -LiteralPath $path)
}

$repoRoot = (& git rev-parse --show-toplevel 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $repoRoot) {
    throw "run this script from inside the Reasonix repository"
}
Set-Location $repoRoot

$currentBranch = (& git symbolic-ref --quiet --short HEAD 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $currentBranch) {
    throw "detached HEAD is not supported"
}
if ((& git status --porcelain).Count -ne 0) {
    throw "working tree must be clean before syncing upstream"
}
if ((Test-GitPath "MERGE_HEAD") -or (Test-GitPath "rebase-merge") -or (Test-GitPath "rebase-apply")) {
    throw "finish or abort the current merge/rebase before syncing upstream"
}
& git show-ref --verify --quiet refs/heads/main-v2
if ($LASTEXITCODE -ne 0) {
    throw "local branch main-v2 does not exist"
}
Invoke-Git remote get-url upstream | Out-Null

$oldRevision = (& git rev-parse main-v2).Trim()
$switched = $false
$newRevision = $null
try {
    Invoke-Git fetch upstream --prune
    if ($currentBranch -ne "main-v2") {
        Invoke-Git switch main-v2
        $switched = $true
    }
    Invoke-Git merge --ff-only upstream/main-v2
    $newRevision = (& git rev-parse main-v2).Trim()
    $upstreamRevision = (& git rev-parse upstream/main-v2).Trim()
    $mainTree = (& git rev-parse "main-v2^{tree}").Trim()
    $upstreamTree = (& git rev-parse "upstream/main-v2^{tree}").Trim()
    if ($newRevision -ne $upstreamRevision -or $mainTree -ne $upstreamTree) {
        throw "main-v2 does not exactly match upstream/main-v2 after synchronization"
    }
}
finally {
    if ($switched) {
        Invoke-Git switch $currentBranch
    }
}

Write-Host "main-v2: $oldRevision -> $newRevision"
if ($oldRevision -ne $newRevision) {
    Write-Host "Overall upstream changes:"
    & git diff --stat "$oldRevision..$newRevision"
    Write-Host "Upstream WebUI changes requiring Baize review:"
    & git diff --stat "$oldRevision..$newRevision" -- internal/serve/index.html internal/serve/login.html internal/serve/serve.go internal/serve/web_assets.go internal/serve/assets
    Write-Host "Upstream workflow changes:"
    & git diff --stat "$oldRevision..$newRevision" -- .github/workflows
}

$known = @{}
foreach ($list in @("scripts/baize-workflows-active.txt", "scripts/baize-workflows-disabled.txt")) {
    foreach ($line in Get-Content -LiteralPath $list) {
        $path = $line.Trim()
        if ($path -and -not $path.StartsWith("#")) { $known[$path] = $true }
    }
}
$unknown = @()
if ($oldRevision -ne $newRevision) {
    foreach ($path in (& git diff --diff-filter=A --name-only "$oldRevision..$newRevision" -- .github/workflows)) {
        if ($path -and -not $known.ContainsKey($path)) { $unknown += $path }
    }
}
if ($unknown.Count -gt 0) {
    Write-Error ("new upstream workflows require an Actions-disabled audit before pushing main-v2:`n  " + ($unknown -join "`n  "))
}

Write-Host "No push or custom merge was performed. Next, inspect the diff and run:"
Write-Host "  git push origin main-v2"
Write-Host "  git switch custom/baize"
Write-Host "  git merge --no-ff --no-commit main-v2"

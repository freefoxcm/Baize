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
& git show-ref --verify --quiet refs/heads/upstream-sync/main-v2
if ($LASTEXITCODE -ne 0) {
    throw "local branch upstream-sync/main-v2 does not exist"
}
Invoke-Git remote get-url upstream | Out-Null

$oldRevision = (& git rev-parse upstream-sync/main-v2).Trim()
$switched = $false
try {
    Invoke-Git fetch upstream --prune
    if ($currentBranch -ne "upstream-sync/main-v2") {
        Invoke-Git switch upstream-sync/main-v2
        $switched = $true
    }
    Invoke-Git merge --ff-only upstream/main-v2
    $newRevision = (& git rev-parse upstream-sync/main-v2).Trim()
}
finally {
    if ($switched) {
        Invoke-Git switch $currentBranch
    }
}

Write-Host "upstream-sync/main-v2: $oldRevision -> $newRevision"
if ($oldRevision -ne $newRevision) {
    & git diff --stat "$oldRevision..$newRevision" -- internal/serve/index.html internal/serve/login.html internal/serve/serve.go
}
Write-Host "No merge or push was performed. Next, inspect the diff and run:"
Write-Host "  git switch custom/baize"
Write-Host "  git merge --no-ff --no-commit upstream-sync/main-v2"

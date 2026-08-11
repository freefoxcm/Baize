param(
    [switch]$Apply,
    [switch]$Check
)

$ErrorActionPreference = "Stop"
if ($Apply -and $Check) { throw "choose either --apply or --check" }
$Gh = if ($env:BAIZE_GH_BIN) { $env:BAIZE_GH_BIN } else { "gh" }

function Invoke-Gh {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & $script:Gh @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "gh $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

$repoRoot = (& git rev-parse --show-toplevel 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $repoRoot) { throw "run this script from inside the Reasonix repository" }
Set-Location $repoRoot

$origin = (& git remote get-url origin).Trim()
if ($origin -notmatch 'github\.com[/:]([^/]+)/(.+?)(?:\.git)?$') {
    throw "origin is not a supported GitHub URL: $origin"
}
$repository = "$($Matches[1])/$($Matches[2])" -replace '\.git$', ''
if ($repository.ToLowerInvariant() -eq 'esengine/deepseek-reasonix') {
    throw "refusing to configure the upstream repository"
}

Invoke-Gh auth status | Out-Null
$admin = (& $Gh api "repos/$repository" --jq '.permissions.admin').Trim()
if ($LASTEXITCODE -ne 0 -or $admin -ne 'true') { throw "GitHub admin permission is required for $repository" }

$activePaths = @(Get-Content scripts/baize-workflows-active.txt | ForEach-Object { $_.Trim() } | Where-Object { $_ -and -not $_.StartsWith('#') })
$disabledPaths = @(Get-Content scripts/baize-workflows-disabled.txt | ForEach-Object { $_.Trim() } | Where-Object { $_ -and -not $_.StartsWith('#') })
$githubManagedPaths = @('dynamic/dependabot/update-graph')

if ($Apply) {
    $currentWorkflowStates = @{}
    foreach ($workflow in @(& $Gh workflow list --repo $repository --all --json path,state | ConvertFrom-Json)) {
        $currentWorkflowStates[$workflow.path] = $workflow.state
    }
    foreach ($path in $disabledPaths) {
        if ($currentWorkflowStates[$path] -eq 'active') {
            Invoke-Gh workflow disable $path --repo $repository
        }
    }
    Invoke-Gh api --method PATCH "repos/$repository" -f default_branch=custom/baize | Out-Null

    foreach ($branch in @('main-v2', 'custom/baize')) {
        $encoded = [Uri]::EscapeDataString($branch)
        $protection = @{
            required_status_checks = $null
            enforce_admins = $false
            required_pull_request_reviews = $null
            restrictions = $null
            required_linear_history = $false
            allow_force_pushes = $false
            allow_deletions = $false
            required_conversation_resolution = $false
        } | ConvertTo-Json -Compress
        $protection | & $Gh api --method PUT "repos/$repository/branches/$encoded/protection" --input - | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "failed to protect $branch" }
    }

    foreach ($path in $activePaths) {
        $enabled = $false
        for ($attempt = 1; $attempt -le 5 -and -not $enabled; $attempt++) {
            & $Gh workflow enable $path --repo $repository
            if ($LASTEXITCODE -eq 0) { $enabled = $true; break }
            Start-Sleep -Seconds 2
        }
        if (-not $enabled) { throw "failed to enable $path after the default-branch change" }
    }
}

$problems = @()
$defaultBranch = (& $Gh repo view $repository --json defaultBranchRef --jq '.defaultBranchRef.name').Trim()
if ($defaultBranch -ne 'custom/baize') { $problems += "default branch is $defaultBranch (expected custom/baize)" }
$actionsEnabled = (& $Gh api "repos/$repository/actions/permissions" --jq '.enabled').Trim()
if ($actionsEnabled -ne 'true') { $problems += "GitHub Actions is disabled for the repository" }

$workflows = @(& $Gh workflow list --repo $repository --all --json path,state | ConvertFrom-Json)
foreach ($workflow in $workflows) {
    if ($workflow.state -eq 'active' -and $workflow.path -notin $activePaths -and $workflow.path -notin $githubManagedPaths) {
        $problems += "unexpected active workflow: $($workflow.path)"
    }
}
foreach ($path in $activePaths) {
    $workflow = $workflows | Where-Object path -eq $path
    if (-not $workflow -or $workflow.state -ne 'active') { $problems += "Baize workflow is not active: $path" }
}
foreach ($branch in @('main-v2', 'custom/baize')) {
    $encoded = [Uri]::EscapeDataString($branch)
    & $Gh api "repos/$repository/branches/$encoded/protection" --silent
    if ($LASTEXITCODE -ne 0) { $problems += "$branch is not protected" }
}

Write-Host "repository: $repository"
Write-Host "default branch: $defaultBranch"
Write-Host "active workflows:"
$workflows | Where-Object state -eq 'active' | Sort-Object path | ForEach-Object { Write-Host "  $($_.path)" }
if ($problems.Count -gt 0) {
    $problems | ForEach-Object { Write-Error $_ -ErrorAction Continue }
    throw "GitHub fork configuration has $($problems.Count) problem(s)"
}
Write-Host "Baize GitHub fork configuration is valid"

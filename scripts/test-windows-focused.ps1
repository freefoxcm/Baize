param()

$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$localAppData = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::LocalApplicationData)
if ([string]::IsNullOrWhiteSpace($localAppData)) {
    throw 'Windows focused tests require a writable LocalApplicationData directory.'
}
# Boot resolves temporary workspaces to their nearest Git root, so test state
# must live outside the checkout while the reusable Go cache stays local.
$runtimeScratchRoot = [System.IO.Path]::GetFullPath((Join-Path $localAppData 'Temp/reasonix-windows-focused'))
$runtimeRoot = [System.IO.Path]::GetFullPath((Join-Path $runtimeScratchRoot "runtime-$PID"))
$goCacheRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot '.windows-focused/go-cache'))
$runtimeScratchPrefix = $runtimeScratchRoot + [System.IO.Path]::DirectorySeparatorChar
if (-not $runtimeRoot.StartsWith($runtimeScratchPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Windows focused test scratch path escaped its root: $runtimeRoot"
}

$temporaryVariables = @('TEMP', 'TMP', 'TMPDIR', 'GOTMPDIR', 'GOCACHE')
$previousEnvironment = @{}
foreach ($name in $temporaryVariables) {
    $previousEnvironment[$name] = [System.Environment]::GetEnvironmentVariable($name, 'Process')
}

Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force $runtimeRoot, $goCacheRoot | Out-Null
    $writeProbe = Join-Path $runtimeRoot '.write-probe'
    [System.IO.File]::WriteAllText($writeProbe, 'ok')
    Remove-Item -LiteralPath $writeProbe -Force

    foreach ($name in @('TEMP', 'TMP', 'TMPDIR', 'GOTMPDIR')) {
        [System.Environment]::SetEnvironmentVariable($name, $runtimeRoot, 'Process')
    }
    [System.Environment]::SetEnvironmentVariable('GOCACHE', $goCacheRoot, 'Process')
    Write-Host "Windows focused tests: temporary files -> $runtimeRoot"
    Write-Host "Windows focused tests: Go build cache -> $goCacheRoot"

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
    foreach ($name in $temporaryVariables) {
        [System.Environment]::SetEnvironmentVariable($name, $previousEnvironment[$name], 'Process')
    }
    if (Test-Path -LiteralPath $runtimeRoot) {
        $resolvedRuntime = [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $runtimeRoot).Path)
        if (-not $resolvedRuntime.StartsWith($runtimeScratchPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Refusing to clean Windows focused test path outside scratch root: $resolvedRuntime"
        }
        Remove-Item -LiteralPath $resolvedRuntime -Recurse -Force
    }
    if ((Test-Path -LiteralPath $runtimeScratchRoot) -and
        -not (Get-ChildItem -LiteralPath $runtimeScratchRoot -Force | Select-Object -First 1)) {
        Remove-Item -LiteralPath $runtimeScratchRoot -Force -ErrorAction SilentlyContinue
    }
    Pop-Location
}

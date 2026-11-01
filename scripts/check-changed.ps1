<#
.SYNOPSIS
  Checks only what changed since a base: format, lint, comment budget, file length and the tests of the changed Go
  packages and frontend folders (Measure build guide §6, issue #M0). Prints only failures; exits 1 when any check fails.
.EXAMPLE
  ./scripts/check-changed.ps1                 # against origin/measure-module
  ./scripts/check-changed.ps1 -Base origin/develop
#>
param([string]$Base = 'origin/measure-module')

$ErrorActionPreference = 'Continue'
$root = Split-Path -Parent $PSScriptRoot
$failures = [System.Collections.Generic.List[string]]::new()

function Invoke-Check([string]$Name, [string]$Dir, [scriptblock]$Command) {
  Push-Location $Dir
  $output = & $Command 2>&1 | Out-String
  $ok = $LASTEXITCODE -eq 0
  Pop-Location
  if (-not $ok) { $failures.Add("== $Name ==`n$($output.Trim())") }
}

# Source files changed since the base, committed or not; generated code is left out.
$changed = @(git -C $root diff --name-only --diff-filter=ACMR "$Base" -- backend frontend) +
  @(git -C $root ls-files --others --exclude-standard -- backend frontend) |
  Where-Object { $_ -and $_ -notmatch 'sqlcdb/|api/generated/' } | Sort-Object -Unique

$goFiles = @($changed | Where-Object { $_ -like 'backend/*.go' })
$webFiles = @($changed | Where-Object { $_ -match '^frontend/(src|e2e)/.*\.(ts|tsx|css)$' })

# Constitution V and VIII: at most 300 non-blank lines, comments at most 5% of them, rounded up.
foreach ($file in @($goFiles + $webFiles)) {
  $lines = @(Get-Content (Join-Path $root $file) | Where-Object { $_.Trim() -ne '' })
  $inBlock = $false
  $comments = 0
  foreach ($line in $lines) {
    $text = $line.Trim()
    if ($inBlock) { $comments++; if ($text.Contains('*/')) { $inBlock = $false }; continue }
    if ($text.StartsWith('/*')) { $comments++; $inBlock = -not $text.Contains('*/'); continue }
    if ($text -match '^//(?!nolint|go:)' -and $text -notmatch 'eslint') { $comments++ }
  }
  $budget = [math]::Ceiling($lines.Count * 0.05)
  if ($comments -gt $budget) { $failures.Add("== Comment budget == $file has $comments comment lines, $budget allowed") }
  if ($lines.Count -gt 300) { $failures.Add("== File length == $file has $($lines.Count) non-blank lines, 300 allowed") }
}

if ($goFiles.Count -gt 0) {
  $packages = @($goFiles | ForEach-Object { './' + (Split-Path ($_ -replace '^backend/', '') -Parent).Replace('\', '/') } |
    Sort-Object -Unique)
  $lint = (Get-Command golangci-lint -ErrorAction SilentlyContinue).Source
  if (-not $lint) { $lint = Join-Path (go env GOPATH) 'bin/golangci-lint.exe' }
  Invoke-Check 'Go format' "$root/backend" { & $lint fmt --diff @packages }
  Invoke-Check 'Go lint' "$root/backend" { & $lint run --concurrency 2 --new-from-rev $Base @packages }
  Invoke-Check 'Go tests' "$root/backend" { go test -short -p 1 @packages }
}

if ($webFiles.Count -gt 0) {
  $relative = @($webFiles | ForEach-Object { $_ -replace '^frontend/', '' })
  $code = @($relative | Where-Object { $_ -notlike '*.css' })
  Invoke-Check 'Prettier' "$root/frontend" { pnpm exec prettier --check @relative }
  if ($code.Count -gt 0) { Invoke-Check 'ESLint' "$root/frontend" { pnpm exec eslint @code } }
  Invoke-Check 'Type check' "$root/frontend" { pnpm run typecheck }
  $tested = @($code | Where-Object { $_ -like 'src/*' })
  if ($tested.Count -gt 0) {
    Invoke-Check 'Vitest (related)' "$root/frontend" { pnpm exec vitest related --run --maxWorkers=1 @tested }
  }
}

if ($failures.Count -gt 0) {
  $failures | ForEach-Object { Write-Output $_; Write-Output '' }
  exit 1
}

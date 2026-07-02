# Сборка в ./bin вместо go run (exe в %TEMP%).
# Kaspersky часто даёт ложное VHO:Trojan.Win64.Gomal.gen на временные Go-бинарники.
# Если срабатывает и здесь — добавьте в исключения Kaspersky:
#   C:\go\STUDY_1\Chapter3\examples\app\bin\
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
$Bin = Join-Path $Root "bin"
$Exe = Join-Path $Bin "processor.exe"

if (-not (Test-Path $Bin)) {
    New-Item -ItemType Directory -Path $Bin | Out-Null
}

go build -o $Exe (Join-Path $Root "cmd/processor")
& $Exe @Args

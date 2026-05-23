param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Tasks
)

$projectRoot = Split-Path -Parent $PSScriptRoot
$env:JAVA_HOME = 'C:\Program Files\Eclipse Adoptium\jdk-17.0.19.10-hotspot'
$env:GRADLE_USER_HOME = Join-Path $env:TEMP 'ticketing-backend-gradle-user-home'

if (-not (Test-Path $env:GRADLE_USER_HOME)) {
    New-Item -ItemType Directory -Path $env:GRADLE_USER_HOME | Out-Null
}

& (Join-Path $projectRoot 'gradlew.bat') @Tasks

exit $LASTEXITCODE
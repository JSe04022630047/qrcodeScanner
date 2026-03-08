@echo off
setlocal

:: Get the short git hash
for /f "tokens=*" %%i in ('git rev-parse --short HEAD') do set GIT_HASH=%%i

echo Building QRScan.exe with commit %GIT_HASH%...

:: Get current date/time
set BUILD_TIME=%date% %time%

:: version, please set before building
set VERSION="r1"

:: Run the build
go build -ldflags "-s -w -H windowsgui -X main.GitCommit=%GIT_HASH% -X 'main.BuildTime=%BUILD_TIME%' -X main.Version=%VERSION%" -o build/QRScan.exe

:: unused, attempting to use fyne's own packaging tool is really confusing. will have to study further
@REM fyne package -os windows -icon icon.png -exe build/QRScan2.exe -options "-ldflags \"-s -w -H windowsgui -X main.GitCommit=%GIT_HASH% -X 'main.BuildTime=%BUILD_TIME%' -X main.Version=%VERSION%\""

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [!] Build FAILED.
    pause
    exit /b %ERRORLEVEL%
)

echo [+] Build successful!
pause
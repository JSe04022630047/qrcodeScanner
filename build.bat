@echo off
setlocal

:: Get the short git hash
for /f "tokens=*" %%i in ('git rev-parse --short HEAD') do set GIT_HASH=%%i

echo Building QRScan.exe with commit %GIT_HASH%...

:: Get current date/time
set BUILD_TIME=%date% %time%

:: version, please set before building
set VERSION="a1"

:: Run the build
go build -ldflags "-X main.GitCommit=%GIT_HASH% -X 'main.BuildTime=%BUILD_TIME%' -X main.Version=%VERSION%" -o build/QRScan.exe

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [!] Build FAILED.
    pause
    exit /b %ERRORLEVEL%
)

echo [+] Build successful!
pause
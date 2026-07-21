@echo off
chcp 65001 >nul 2>nul
setlocal

REM ============================================================
REM  StudentHub - Stop All Services
REM  Kills processes on ports 8088 (backend) and 5173 (frontend)
REM ============================================================

echo.
echo Stopping StudentHub services...

for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":8088 " ^| findstr "LISTENING" 2^>nul') do (
    echo   Killing backend PID %%a
    taskkill /F /PID %%a >nul 2>nul
)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":5173 " ^| findstr "LISTENING" 2^>nul') do (
    echo   Killing frontend PID %%a
    taskkill /F /PID %%a >nul 2>nul
)

echo   Done.
timeout /t 2 /nobreak >nul
exit

@echo off
chcp 65001 >nul 2>nul
setlocal enabledelayedexpansion

REM ============================================================
REM  StudentHub - One-click Start (Backend + Frontend)
REM  Starts Spring Boot backend on 8088, then Vite frontend on 5173.
REM  Press Ctrl+C in each window to stop.
REM ============================================================

set "ROOT=d:\Developing\WorkSpace\StudentManagementSystemVD"

echo.
echo ============================================================
echo  StudentHub One-Click Start
echo ============================================================

REM ---- Step 1: Prepare directories ----
echo [1/4] Preparing directories...
cd /d "%ROOT%"
if not exist "backend\data" mkdir "backend\data"
if not exist "backend\logs" mkdir "backend\logs"
echo   OK.

REM ---- Step 2: Kill any process on port 8088 and 5173 ----
echo [2/4] Cleaning up old processes on ports 8088 and 5173...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":8088 " ^| findstr "LISTENING" 2^>nul') do (
    echo   Killing PID %%a on port 8088
    taskkill /F /PID %%a >nul 2>nul
)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":5173 " ^| findstr "LISTENING" 2^>nul') do (
    echo   Killing PID %%a on port 5173
    taskkill /F /PID %%a >nul 2>nul
)
echo   OK.

REM ---- Step 3: Build backend (force recompile) ----
echo [3/4] Compiling backend...
cd /d "%ROOT%\backend"
call mvn -q compile
if errorlevel 1 (
    echo.
    echo ============================================================
    echo  [ERROR] Backend compilation FAILED!
    echo  Fix the errors above, then run this script again.
    echo ============================================================
    pause
    exit /b 1
)
echo   Backend compiled OK.

REM ---- Step 4: Start backend and frontend in new windows ----
echo [4/4] Starting services...

REM Start backend in a new window (visible logs)
start "StudentHub Backend (port 8088)" cmd /k "cd /d %ROOT%\backend && title StudentHub Backend (port 8088) && mvn spring-boot:run"

REM Wait for backend to be ready (max 60 seconds)
echo   Waiting for backend to start...
set "READY=0"
for /l %%i in (1,1,60) do (
    if "!READY!"=="0" (
        timeout /t 1 /nobreak >nul
        powershell -NoProfile -Command "try { $r = Invoke-WebRequest -Uri 'http://127.0.0.1:8088/api/v1/actuator/health' -TimeoutSec 2 -UseBasicParsing; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" >nul 2>nul
        if !errorlevel! equ 0 (
            set "READY=1"
            echo   Backend is ready! ^(took %%i seconds^)
        )
    )
)

if "!READY!"=="0" (
    echo.
    echo ============================================================
    echo  [WARNING] Backend did not respond within 60 seconds.
    echo  Check the "StudentHub Backend" window for errors.
    echo  Starting frontend anyway...
    echo ============================================================
) else (
    REM Start frontend in a new window (use npm since pnpm may not be installed)
    start "StudentHub Frontend (port 5173)" cmd /k "cd /d %ROOT%\frontend && title StudentHub Frontend (port 5173) && npm run dev"
    echo   Frontend started.
)

echo.
echo ============================================================
echo  Startup complete!
echo.
echo  Backend  : http://127.0.0.1:8088/api/v1  ^(window: StudentHub Backend^)
echo  Frontend : http://127.0.0.1:5173/         ^(window: StudentHub Frontend^)
echo  Login    : admin / admin@123
echo.
echo  To stop: close the two popup windows, or run stop_all.bat
echo ============================================================
echo.
echo This window will close in 5 seconds...
timeout /t 5 /nobreak >nul
exit

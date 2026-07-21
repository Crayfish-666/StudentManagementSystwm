@echo off
chcp 65001 >nul 2>nul
setlocal

REM ============================================================
REM  StudentHub Backend - Build and Run (Java Spring Boot)
REM  Compiles the backend and starts it with visible logs.
REM ============================================================

cd /d "d:\Developing\WorkSpace\StudentManagementSystemVD\backend"

echo.
echo [1/3] Ensuring data directory exists...
if not exist "data" mkdir data
if not exist "logs" mkdir logs
echo   data/ and logs/ ready.

echo.
echo [2/3] Compiling backend (mvn compile)...
call mvn -q compile
if errorlevel 1 (
    echo.
    echo [ERROR] Compilation failed! Fix the errors above before running.
    pause
    exit /b 1
)
echo   Compilation OK.

echo.
echo [3/3] Starting Spring Boot...
echo   Port: 8088
echo   Context: /api/v1
echo   Press Ctrl+C to stop.
echo ============================================================
echo.
call mvn spring-boot:run
pause

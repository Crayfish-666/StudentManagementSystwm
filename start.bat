@echo off
title StudentHub Launcher

echo =================================================================
echo             StudentHub Management System 
echo               One-Click Parallel Launcher
echo =================================================================
echo.

echo [1/2] Starting Spring Boot 3 Backend Server (Port :8088)...
start "Backend" cmd /c "cd /d "%~dp0backend" && mvn spring-boot:run"

echo [2/2] Starting Vue 3 + Vite 5 Frontend Dev Server (Port :5173)...
start "Frontend" cmd /c "cd /d "%~dp0frontend" && npm run dev"

echo.
echo =================================================================
echo  [SUCCESS] Both Backend and Frontend servers launched!
echo  - Frontend Web UI: http://127.0.0.1:5173/
echo  - Backend API URL: http://127.0.0.1:8088/api/v1
echo =================================================================
echo.

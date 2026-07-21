@echo off
chcp 65001 >nul
title StudentHub 一键启动程序

echo =================================================================
echo             StudentHub 学生一站式自主管理系统 
echo                      一键并行启动脚本
echo =================================================================
echo.

:: 1. 检查 Java 环境
where java >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Java 环境，请确保 Java 21+ 已安装并配置环境变量！
    pause
    exit /b 1
)

:: 2. 检查 Node.js 环境
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Node.js 环境，请确保 Node.js 18+ 已安装！
    pause
    exit /b 1
)

echo [1/2] 正在启动 Spring Boot 3 后端服务 (端口 :8088)...
start "StudentHub Backend (:8088)" cmd /k "cd /d %~dp0backend && mvn spring-boot:run"

echo [2/2] 正在启动 Vue 3 + Vite 5 前端开发服务器 (端口 :5173)...
start "StudentHub Frontend (:5173)" cmd /k "cd /d %~dp0frontend && npm run dev"

echo.
echo =================================================================
echo  [启动成功] 双端服务已在独立的终端窗口中并行启动：
echo  - 前端 WEB 界面: http://127.0.0.1:5173/
echo  - 后端 API 服务: http://127.0.0.1:8088/api/v1
echo  - 数据库 WAL 文件: ./backend/data/studenthub.db
echo =================================================================
echo.
pause

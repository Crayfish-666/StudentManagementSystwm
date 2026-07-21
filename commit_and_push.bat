@echo off
chcp 65001 >nul 2>nul
setlocal

REM ============================================================
REM  StudentHub - Commit and Push
REM ============================================================

set "REPO_URL=https://github.com/Crayfish-666/StudentManagementSystwm.git"
set "ROOT=d:\Developing\WorkSpace\StudentManagementSystemVD"

cd /d "%ROOT%"

REM Check git repo
if not exist ".git" (
    echo [INFO] Initializing git repository...
    git init
    git branch -M main
)

REM Configure remote
echo [1/4] Configuring remote origin...
git remote remove origin >nul 2>nul
git remote add origin %REPO_URL%
echo   OK.

REM Stage all changes
echo [2/4] Staging changes...
git add -A
echo   OK.

REM Commit with UTF-8 message file
echo [3/4] Creating commit...
echo feat(cmp): patch CMP module for ranking data> "%ROOT%\commit_msg.txt"
echo.>> "%ROOT%\commit_msg.txt"
echo - CmpModuleController: rewrite /cmp/scores with full field mapping>> "%ROOT%\commit_msg.txt"
echo   (student_no, college_name, college_class_name, academic_year, computed_at,>> "%ROOT%\commit_msg.txt"
echo   rank_in_class, rank_in_college) using subqueries instead of window functions>> "%ROOT%\commit_msg.txt"
echo - CmpModuleController: add /cmp/scores/me, /cmp/scores/{id}, /cmp/scores/me/history,>> "%ROOT%\commit_msg.txt"
echo   /cmp/scores/{id}/recompute, /cmp/scores/compute endpoints>> "%ROOT%\commit_msg.txt"
echo - CmpModuleController: add /cmp/dashboard/{kpi,trends,distribution,>> "%ROOT%\commit_msg.txt"
echo   active-assoc-by-college,incident-level} endpoints>> "%ROOT%\commit_msg.txt"
echo - CmpModuleController: add /cmp/rule-versions CRUD endpoints>> "%ROOT%\commit_msg.txt"
echo - cmp/Dashboard.vue: replace Promise.all with Promise.allSettled for fault tolerance>> "%ROOT%\commit_msg.txt"
echo - V1.2__seed_dashboard_data.sql: add sys_role, sys_user_role, sys_dict,>> "%ROOT%\commit_msg.txt"
echo   st_activity, st_recruit_plan, sq_incident, qg_difficulty_cert, file_meta,>> "%ROOT%\commit_msg.txt"
echo   cmp_ai_evaluation, sys_menu seed data>> "%ROOT%\commit_msg.txt"
echo - run_backend.bat: fix port number (8088)>> "%ROOT%\commit_msg.txt"
echo - start_all.bat: use npm instead of pnpm (system default)>> "%ROOT%\commit_msg.txt"
echo - vite.config.js: change proxy target from localhost to 127.0.0.1>> "%ROOT%\commit_msg.txt"
echo - application.yml: unify port to 8088, remove spring.ai.openai config>> "%ROOT%\commit_msg.txt"
git commit -F "%ROOT%\commit_msg.txt"
del "%ROOT%\commit_msg.txt" >nul 2>nul
echo   OK.

REM Push
echo [4/4] Pushing to GitHub...
git push -u origin main
if errorlevel 1 (
    echo.
    echo ============================================================
    echo  [ERROR] Push failed!
    echo.
    echo  Common causes:
    echo    1. Not authenticated - run: git config --global credential.helper manager
    echo       First push will open browser for GitHub login
    echo    2. Remote repo does not exist - create empty repo at:
    echo       https://github.com/Crayfish-666/StudentManagementSystwm
    echo    3. Network issue - check proxy or retry
    echo ============================================================
    pause
    exit /b 1
)

echo.
echo ============================================================
echo  Push successful!
echo  Repo: %REPO_URL%
echo ============================================================
pause
exit

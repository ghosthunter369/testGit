@echo off
echo ========================================
echo   Claw Backend + Frontend Startup
echo ========================================
echo.

:: Check Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please install Go from https://go.dev/dl/
    pause
    exit /b 1
)

:: Check Node
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed or not in PATH
    pause
    exit /b 1
)

echo [1/4] Checking MySQL and Redis...
echo Make sure MySQL (3306) and Redis (6379) are running!
echo.

echo [2/4] Starting Backend (port 8080)...
start "Claw Backend" cmd /k "cd /d C:\goProj\backend && go mod tidy && go run main.go"
timeout /t 3 /nobreak >nul

echo [3/4] Starting Frontend (port 5173)...
start "Claw Frontend" cmd /k "cd /d C:\goProj\backend\frontend && npm run dev"
timeout /t 3 /nobreak >nul

echo [4/4] Opening browser...
timeout /t 5 /nobreak >nul
start http://localhost:5173

echo.
echo ========================================
echo   Servers started!
echo   Backend:  http://localhost:8080
echo   Frontend: http://localhost:5173
echo ========================================
echo.
pause

@echo off
REM 构建脚本 - Windows

echo ==================================
echo   构建供应商管理系统
echo ==================================
echo.

REM 1. 构建前端
echo [1/2] 构建前端...
cd frontend
call npm install
call npm run build
cd ..

REM 2. 移动前端构建产物到后端
echo [2/2] 移动前端文件到后端...
if not exist "backend\frontend" mkdir backend\frontend
xcopy /E /I /Y "frontend\dist" "backend\frontend\dist"

echo.
echo ==================================
echo   构建完成！
echo ==================================
echo.
echo 运行: cd backend ^&^& go run main.go
echo.
pause

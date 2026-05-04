@echo off
REM 构建脚本 - Windows

echo ==================================
echo   构建供应商管理系统
echo ==================================
echo.

REM 1. 构建前端
echo [1/3] 构建前端...
cd frontend
call npm install
call npm run build
cd ..

REM 2. 移动前端构建产物到后端
echo [2/3] 移动前端文件到后端...
if not exist "backend\frontend" mkdir backend\frontend
xcopy /E /I /Y "frontend\dist" "backend\frontend\dist"

REM 3. 编译后端
echo [3/3] 编译后端...
if not exist "build" mkdir build
cd backend
go build -ldflags="-s -w" -o ..\build\supplier-card-manager.exe .
cd ..

echo.
echo ==================================
echo   构建完成！
echo ==================================
echo   产物: build\supplier-card-manager.exe
echo.
echo 运行: build\supplier-card-manager.exe
echo.
pause

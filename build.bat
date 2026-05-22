@echo off

echo [1/3] 前端...
cd frontend
call npm install >nul 2>&1
call npm run build
cd ..

echo [2/3] 复制...
if not exist "backend\frontend" mkdir backend\frontend
xcopy /E /I /Y /Q "frontend\dist" "backend\frontend\dist" >nul

echo [3/3] 后端...
if not exist "build" mkdir build
cd backend
go build -ldflags="-s -w" -o ..\build\supplier-card-manager.exe .
cd ..

echo 完成: build\supplier-card-manager.exe

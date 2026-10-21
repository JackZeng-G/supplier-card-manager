@echo off

echo [1/3] 前端...
cd frontend
call pnpm install --frozen-lockfile >nul 2>&1
call pnpm run build
cd ..

echo [2/3] 复制...
if not exist "backend\frontend" mkdir backend\frontend
xcopy /E /I /Y /Q "frontend\dist" "backend\frontend\dist" >nul

echo [3/3] 后端...
cd backend
go build -ldflags="-s -w" -o ..\build\supplier-card-manager.exe .
cd ..

echo 完成: build\supplier-card-manager.exe

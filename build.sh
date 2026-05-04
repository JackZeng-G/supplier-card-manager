#!/bin/bash
# 构建脚本 - Linux/Mac

echo "=================================="
echo "  构建供应商管理系统"
echo "=================================="
echo

# 1. 构建前端
echo "[1/3] 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 2. 移动前端构建产物到后端
echo "[2/3] 移动前端文件到后端..."
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/

# 3. 编译后端
echo "[3/3] 编译后端..."
mkdir -p dist
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../dist/supplier-card-manager .
cd ..

echo
echo "=================================="
echo "  构建完成！"
echo "=================================="
echo "  产物: dist/supplier-card-manager"
echo
echo "运行: cd backend && ../dist/supplier-card-manager"
echo

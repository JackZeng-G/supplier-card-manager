#!/bin/bash
# 构建脚本 - Linux/Mac

echo "=================================="
echo "  构建供应商管理系统"
echo "=================================="
echo

# 1. 构建前端
echo "[1/2] 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 2. 移动前端构建产物到后端
echo "[2/2] 移动前端文件到后端..."
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/

echo
echo "=================================="
echo "  构建完成！"
echo "=================================="
echo
echo "运行: cd backend && go run main.go"
echo

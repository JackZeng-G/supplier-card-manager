#!/bin/bash
# 构建脚本 - Linux/Mac（从项目根目录执行）

set -e

echo "=================================="
echo "  构建供应商名片管理系统"
echo "=================================="
echo

# 1. 构建前端
echo "[1/3] 构建前端..."
npm --prefix frontend install
npm --prefix frontend run build

# 2. 移动前端构建产物到后端（go:embed 需要）
echo "[2/3] 复制前端产物到后端..."
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/

# 3. 编译后端
echo "[3/3] 编译后端..."
mkdir -p build
GOOS=linux GOARCH=amd64 go build -C backend -ldflags="-s -w" -o ../build/supplier-card-manager .

echo
echo "=================================="
echo "  构建完成！"
echo "=================================="
echo "  产物: build/supplier-card-manager"
echo
echo "运行: ./build/supplier-card-manager"

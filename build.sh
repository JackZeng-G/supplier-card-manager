#!/bin/bash
# 构建脚本 - Linux/Mac

set -e

echo "[1/3] 前端..."
cd frontend
npm install
npm run build
cd ..

echo "[2/3] 复制..."
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/

echo "[3/3] 后端..."
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../build/supplier-card-manager .
cd ..

echo "完成: build/supplier-card-manager"

#!/bin/bash
set -e

echo "[1/3] 前端..."
npm --prefix frontend install
npm --prefix frontend run build

echo "[2/3] 复制..."
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/

echo "[3/3] 后端..."
mkdir -p build
GOOS=linux GOARCH=amd64 go build -C backend -ldflags="-s -w" -o ../build/supplier-card-manager .

echo "完成: build/supplier-card-manager"

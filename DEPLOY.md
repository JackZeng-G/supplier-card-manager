# Docker 部署指南

## 快速开始

### 1. 准备环境变量

编辑 `.env` 文件或设置系统环境变量：

```bash
# Windows PowerShell
$env:TENCENT_SECRET_ID="your_secret_id"
$env:TENCENT_SECRET_KEY="your_secret_key"

# Linux/Mac
export TENCENT_SECRET_ID="your_secret_id"
export TENCENT_SECRET_KEY="your_secret_key"
```

### 2. 本地构建二进制

```bash
# Linux/Mac
./build.sh

# Windows
build.bat
```

产物位于 `build/supplier-card-manager`（Linux）或 `build/supplier-card-manager.exe`（Windows）。

### 3. Docker 部署

```bash
# 构建并启动
docker-compose up -d --build

# 或使用快速启动脚本
./start-docker.sh        # Linux/Mac
start-docker.bat         # Windows
```

### 4. 访问应用

- 应用地址: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 常用命令

```bash
# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps

# 停止服务
docker-compose down

# 停止并删除数据卷（会删除所有数据）
docker-compose down -v

# 重新构建并启动
docker-compose up -d --build

# 进入容器
docker-compose exec app sh

# 备份数据库
docker run --rm -v supplier-card-manager_app-data:/app/data -v $(pwd):/backup alpine tar czf /backup/supplier-data-backup.tar.gz /app/data
```

## 数据持久化

Docker volume `app-data` 挂载到容器内 `/app/data`，包含：
- SQLite 数据库
- 临时上传文件
- 永久图片存储

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| SERVER_PORT | 8080 | 服务端口 |
| DATABASE_PATH | ./data/suppliers.db | 数据库路径 |
| TEMP_UPLOAD_PATH | ./data/uploads | 临时上传目录 |
| IMAGE_PATH | ./data/images | 图片存储目录 |
| TENCENT_SECRET_ID | - | 腾讯云 OCR SecretId |
| TENCENT_SECRET_KEY | - | 腾讯云 OCR SecretKey |
| TENCENT_REGION | ap-guangzhou | 腾讯云 OCR 区域 |

## 生产环境部署建议

### 1. 使用反向代理

推荐使用 Nginx 作为反向代理：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 2. 配置 HTTPS

使用 Let's Encrypt 和 Certbot 获取免费 SSL 证书：

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### 3. 定期备份数据

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker run --rm \
  -v supplier-card-manager_app-data:/app/data \
  -v /backup:/backup \
  alpine tar czf /backup/supplier-$DATE.tar.gz /app/data
```

## 故障排查

### 图片无法显示

检查 volumes 权限：
```bash
docker-compose exec app ls -la /app/data/images
```

### 数据库连接失败

检查健康检查：
```bash
curl http://localhost:8080/health
docker-compose logs --tail=50 app
```

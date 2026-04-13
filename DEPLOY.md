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

### 2. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps
```

### 3. 访问应用

- 前端: http://localhost
- 后端API: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 常用命令

```bash
# 停止服务
docker-compose down

# 停止并删除数据卷（⚠️ 会删除所有数据）
docker-compose down -v

# 重新构建并启动
docker-compose up -d --build

# 查看后端日志
docker-compose logs -f backend

# 查看前端日志
docker-compose logs -f frontend

# 进入后端容器
docker-compose exec backend sh

# 备份数据库
docker run --rm -v supplier-backend-data:/data -v $(pwd):/backup alpine tar czf /backup/supplier-data-backup.tar.gz /data
```

## 数据持久化

Docker volumes 用于持久化数据：
- `backend-data`: SQLite 数据库
- `backend-uploads`: 临时上传文件
- `backend-images`: 永久图片存储

## 生产环境部署建议

### 1. 使用反向代理

推荐使用 Nginx 或 Traefik 作为反向代理：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 2. 配置HTTPS

使用 Let's Encrypt 和 Certbot 获取免费SSL证书：

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### 3. 设置防火墙

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 4. 定期备份数据

创建备份脚本：

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker run --rm \
  -v supplier-backend-data:/data \
  -v /backup:/backup \
  alpine tar czf /backup/supplier-$DATE.tar.gz /data
```

## 故障排查

### 前端无法访问后端

检查网络连接：
```bash
docker-compose exec frontend ping backend
```

### 图片无法显示

检查volumes权限：
```bash
docker-compose exec backend ls -la /app/images
```

### 数据库连接失败

检查数据目录：
```bash
docker-compose exec backend ls -la /app/data
```

### 查看详细错误日志

```bash
docker-compose logs --tail=100 backend
docker-compose logs --tail=100 frontend
```

## 端口说明

| 服务 | 容器端口 | 主机端口 |
|-----|---------|---------|
| 前端 | 80 | 80 |
| 后端 | 8080 | 8080 |

如果需要修改端口，编辑 `docker-compose.yml` 中的 `ports` 配置。

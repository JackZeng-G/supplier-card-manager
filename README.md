# 🎴 Supplier Card Manager

<div align="center">

**供应商名片管理系统** — 基于 OCR 的智能名片识别与供应商档案管理

[![Vue 3](https://img.shields.io/badge/Vue%203-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-GPL%20v3-blue?style=for-the-badge)](LICENSE)

</div>

---

## ✨ 功能特性

| 功能 | 描述 |
|------|------|
| 📸 **OCR 智能识别** | 支持名片正反面自动识别，提取公司、联系人、电话等信息 |
| 📋 **供应商管理** | 完整的增删改查，支持合作状态追踪（合作中/待开发/已暂停） |
| 🔍 **多维度筛选** | 按公司名称、联系人、状态、运输方式快速搜索 |
| 📊 **数据统计** | 实时统计供应商数量及各状态分布 |
| 📤 **Excel 导出** | 一键导出所有供应商数据到 Excel |
| 🐳 **Docker 部署** | 一键启动，无需配置复杂环境 |

---

## 🚀 快速开始

### 方式一：Docker 部署（推荐）

```bash
# 1. 克隆项目
git clone http://zqy.x64.baby:20011/zqyhqw/supplier-card-manager.git
cd supplier-card-manager

# 2. 配置腾讯云 OCR（可选，用于名片识别）
cp .env.example .env
# 编辑 .env 填入你的腾讯云密钥

# 3. 启动服务
docker-compose up -d

# 4. 访问应用
# 前端: http://localhost
# API:  http://localhost:8080
```

### 方式二：本地开发

**后端启动**
```bash
cd backend

# 设置环境变量（可选）
export TENCENT_SECRET_ID=your_secret_id
export TENCENT_SECRET_KEY=your_secret_key

# 运行
go run main.go
```

**前端启动**
```bash
cd frontend
npm install
npm run dev
```

---

## ⚙️ 环境配置

### 腾讯云 OCR 配置（推荐）

> 未配置时将返回模拟数据，适合演示和测试

1. [注册腾讯云账号](https://cloud.tencent.com)
2. [开通智能结构化 OCR](https://console.cloud.tencent.com/ocr/overview)
3. [创建 API 密钥](https://console.cloud.tencent.com/cam/capi)
4. 配置环境变量：

```bash
TENCENT_SECRET_ID=your_secret_id
TENCENT_SECRET_KEY=your_secret_key
```

### 完整配置项

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 后端服务端口 | `8080` |
| `DATABASE_PATH` | SQLite 数据库路径 | `./data/suppliers.db` |
| `UPLOAD_PATH` | 上传文件存储路径 | `./uploads` |
| `IMAGES_PATH` | 永久图片存储路径 | `./images` |
| `TENCENT_SECRET_ID` | 腾讯云 SecretId | - |
| `TENCENT_SECRET_KEY` | 腾讯云 SecretKey | - |

---

## 📁 项目结构

```
supplier-card-manager/
├── 📁 backend/              # Go 后端
│   ├── 📁 config/          # 配置管理
│   ├── 📁 handlers/        # HTTP 接口处理器
│   ├── 📁 middleware/      # Gin 中间件
│   ├── 📁 models/          # 数据模型 & 数据库
│   ├── 📁 services/        # 业务逻辑（OCR、图片处理）
│   ├── 📄 main.go          # 程序入口
│   └── 📄 Dockerfile
│
├── 📁 frontend/             # Vue 3 前端
│   ├── 📁 src/
│   │   ├── 📁 api/         # API 接口封装
│   │   ├── 📁 router/      # 路由配置
│   │   ├── 📁 views/       # 页面组件
│   │   │   ├── List.vue    # 供应商列表
│   │   │   ├── Detail.vue  # 详情/编辑
│   │   │   └── Upload.vue  # 名片上传
│   │   └── 📄 App.vue      # 应用布局
│   └── 📄 package.json
│
├── 📁 名片范例/             # 示例名片图片
├── 📄 docker-compose.yml    # Docker 编排配置
├── 📄 README.md            # 本文件
└── 📄 DEPLOY.md            # 详细部署指南
```

---

## 🔌 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/suppliers/upload` | 上传名片正面识别 |
| `POST` | `/api/suppliers/upload-back` | 上传名片反面识别 |
| `GET` | `/api/suppliers` | 获取供应商列表（支持分页、搜索、筛选） |
| `GET` | `/api/suppliers/:id` | 获取单个供应商详情 |
| `POST` | `/api/suppliers` | 创建供应商 |
| `PUT` | `/api/suppliers/:id` | 更新供应商信息 |
| `DELETE` | `/api/suppliers/:id` | 删除供应商 |
| `GET` | `/api/suppliers/export` | 导出 Excel |
| `GET` | `/api/suppliers/stats` | 获取统计数据 |

---

## 🛠️ 技术栈

### 后端
- **Go 1.21+** — 高性能后端语言
- **Gin** — 轻量级 Web 框架
- **GORM** — 强大的 ORM 库
- **SQLite** — 零配置嵌入式数据库
- **腾讯云 OCR** — 智能文字识别

### 前端
- **Vue 3** — 渐进式 JavaScript 框架
- **Element Plus** — 企业级 UI 组件库
- **Vue Router** — 官方路由管理器
- **Axios** — HTTP 客户端
- **Vite** — 下一代前端构建工具

---

## 🐳 Docker 常用命令

```bash
# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重建并启动
docker-compose up -d --build

# 备份数据
docker run --rm -v supplier-backend-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/supplier-backup-$(date +%Y%m%d).tar.gz /data
```

详见 [DEPLOY.md](DEPLOY.md)

---

## 📌 注意事项

1. **单人使用**：当前版本为单人版，无需登录
2. **数据持久化**：Docker 部署时数据存储在 volumes 中
3. **图片存储**：上传的名片图片会保存到 `images/` 目录
4. **OCR 识别**：需要联网调用腾讯云 API

---

## 📜 License

GNU General Public License v3.0 © 2026 JackZeng

本项目采用 GPLv3 许可证，您可以自由使用、修改和分发，但任何衍生作品必须同样采用 GPLv3 许可证并保持开源。

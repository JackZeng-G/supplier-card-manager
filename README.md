# 供应商名片管理系统

一个基于 Go + Vue 3 的供应商名片管理系统，支持名片OCR识别、数据管理和展示。

## 功能特性

- 名片图片上传与OCR自动识别
- 供应商信息增删改查
- 数据搜索、筛选、分页
- Excel导出功能
- 单人版本，无需登录

## 技术栈

### 后端
- Go 1.21+
- Gin (Web框架)
- GORM (ORM)
- SQLite (数据库)
- 腾讯云OCR (文字识别)

### 前端
- Vue 3
- Element Plus
- Vue Router
- Axios
- Vite

## 快速开始

### 后端启动

```bash
cd backend

# 设置腾讯云OCR密钥（可选，不设置将返回模拟数据）
set TENCENT_SECRET_ID=your_secret_id
set TENCENT_SECRET_KEY=your_secret_key

# 运行
go run main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
```

前端服务将在 `http://localhost:3000` 启动。

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_PORT | 后端服务端口 | 8080 |
| DATABASE_PATH | SQLite数据库路径 | ./data/suppliers.db |
| UPLOAD_PATH | 上传文件存储路径 | ./uploads |
| TENCENT_SECRET_ID | 腾讯云SecretId | 空 |
| TENCENT_SECRET_KEY | 腾讯云SecretKey | 空 |

### 腾讯云OCR配置

1. 注册腾讯云账号：https://cloud.tencent.com
2. 开通OCR服务：https://console.cloud.tencent.com/ocr
3. 创建密钥：https://console.cloud.tencent.com/cam/capi
4. 设置环境变量

## API接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/suppliers/upload | 上传名片并OCR识别 |
| GET | /api/suppliers | 获取供应商列表 |
| GET | /api/suppliers/:id | 获取单个供应商 |
| POST | /api/suppliers | 创建供应商 |
| PUT | /api/suppliers/:id | 更新供应商 |
| DELETE | /api/suppliers/:id | 删除供应商 |
| GET | /api/suppliers/export | 导出Excel |

## 项目结构

```
供应商记录/
├── backend/                 # 后端代码
│   ├── main.go             # 入口文件
│   ├── config/             # 配置管理
│   ├── models/             # 数据模型
│   ├── handlers/           # 接口处理
│   ├── services/           # 业务服务
│   ├── middleware/         # 中间件
│   ├── uploads/            # 上传文件
│   └── data/               # 数据库文件
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── views/          # 页面组件
│   │   ├── api/            # API封装
│   │   └── router/         # 路由配置
│   └── ...
└── 名片范例/               # 示例名片图片
```

## License

MIT
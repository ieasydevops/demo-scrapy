# 政府采购网监控系统

政府采购网信息监控与推送系统，支持多网页监控、关键词过滤、定时采集和邮件推送功能。

## 快速开始

### Docker 部署（推荐）

```bash
# 1. 克隆项目
git clone <repository-url>
cd demo-scrapy

# 2. 配置邮件信息
vim config.yaml  # 修改 smtp_user 和 smtp_pass

# 3. 启动服务
make docker-build
make docker-run

# 4. 访问
# 前端: http://localhost
# 后端 API: http://localhost:5080
# Swagger: http://localhost/swagger/index.html
```

### 本地开发

```bash
# 后端
make build
./bin/server

# 前端（新终端）
cd frontend
npm install
npm run serve
```

## 目录

- [系统需求](#系统需求)
- [系统架构](#系统架构)
- [核心功能](#核心功能)
- [系统配置](#系统配置)
- [部署方案](#部署方案)
  - [macOS 部署](#macos-部署)
  - [Windows 部署](#windows-部署)
  - [Linux 部署](#linux-部署)
  - [Docker 部署](#docker-部署)
- [开发指南](#开发指南)
- [API 文档](#api-文档)

## 系统需求

### 功能需求

1. **网页监控**
   - 支持多个政府采购网站监控
   - 可配置监控网页列表
   - 支持关键词过滤

2. **定时采集**
   - 支持定时采集任务配置
   - 可设置采集时间和频率（每日/每小时）
   - 自动去重，避免重复采集

3. **邮件推送**
   - 支持多个订阅邮箱
   - 可配置推送时间
   - 自动发送新公告到订阅邮箱

4. **Web 管理界面**
   - 网页列表管理
   - 关键词管理
   - 监控配置管理
   - 订阅配置管理
   - 采购信息动态展示

### 技术需求

- **后端**: Go 1.22+
- **前端**: Vue 3 + Element Plus
- **数据库**: SQLite
- **Web 框架**: Gin
- **定时任务**: Cron
- **邮件服务**: SMTP

## 系统架构

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        前端层 (Vue 3)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 网页管理 │  │ 关键词管理│  │监控配置  │  │订阅配置  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           采购信息动态展示 (时间排序)                  │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP/REST API
┌──────────────────────▼──────────────────────────────────────┐
│                     后端层 (Go + Gin)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  API 路由    │  │  配置管理    │  │  数据库层     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              定时任务调度器 (Cron)                     │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │   │
│  │  │ 采集任务 │  │ 邮件推送 │  │ 数据存储 │            │   │
│  │  └──────────┘  └──────────┘  └──────────┘            │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              网页爬虫模块 (GoQuery)                   │   │
│  │  - HTML 解析采集                                      │   │
│  │  - API 搜索采集                                       │   │
│  │  - 关键词过滤                                         │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    数据层 (SQLite)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 网页表   │  │ 关键词表 │  │监控配置表│  │订阅配置表│   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   公告信息表                           │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 模块说明

1. **前端模块** (Vue 3)
   - 使用 Element Plus UI 组件库
   - 通过 Axios 调用后端 API
   - 支持响应式布局

2. **后端 API 模块** (Go + Gin)
   - RESTful API 设计
   - Swagger 文档支持
   - CORS 跨域支持

3. **定时任务模块** (Cron)
   - 基于 robfig/cron 实现
   - 支持动态任务加载
   - 可配置采集时间和频率

4. **爬虫模块** (GoQuery)
   - HTML 解析采集
   - API 搜索采集
   - 关键词过滤

5. **邮件模块** (SMTP)
   - 支持 SMTP 邮件发送
   - 邮件模板格式化
   - 批量发送支持

6. **数据存储** (SQLite)
   - 轻量级数据库
   - 支持事务操作
   - 数据持久化

## 核心功能

### 1. 网页监控流程

```
开始
  │
  ├─> 读取监控配置
  │     │
  │     ├─> 获取网页列表
  │     ├─> 获取关键词列表
  │     └─> 获取采集时间配置
  │
  ├─> 定时触发采集任务
  │     │
  │     ├─> 访问目标网页
  │     ├─> 解析 HTML/调用 API
  │     ├─> 关键词过滤
  │     └─> 提取公告信息
  │
  ├─> 数据去重处理
  │     │
  │     ├─> 检查 URL 是否已存在
  │     └─> 保存新公告到数据库
  │
  ├─> 获取新公告列表
  │
  └─> 发送邮件通知
        │
        └─> 完成
```

### 2. 定时任务机制

- **采集任务**: 每10分钟执行一次（可配置）
- **邮件推送**: 根据订阅配置的时间执行（如每天17:00）
- **任务管理**: 支持动态添加/删除任务

### 3. 数据流程

```
外部网站
  │
  ▼
爬虫模块 (采集数据)
  │
  ▼
数据验证 (关键词过滤、去重)
  │
  ▼
SQLite 数据库 (存储)
  │
  ▼
API 接口 (查询)
  │
  ▼
前端展示 / 邮件推送
```

## 系统配置

### 配置文件说明

配置文件: `config.yaml`

```yaml
# 监控网页列表
web_pages:
  - name: 深圳政府采购网
    url: http://zfcg.szggzy.com:8081/gsgg/secondPage.html

# 关键词列表
keywords:
  - 生态环境局

# 监控配置
monitor_configs:
  - web_page_name: 深圳政府采购网
    crawl_time: "9"        # 采集时间（小时，0-23）
    crawl_freq: daily      # 采集频率：daily/hourly
    keywords:
      - 生态环境局

# 邮件配置
email:
  smtp_host: smtp.qq.com
  smtp_user: your_email@qq.com
  smtp_pass: your_smtp_password

# 服务器配置
server:
  port: 5080              # API 服务端口
  db_path: ./monitor.db  # 数据库路径
```

### 环境变量

- `DB_PATH`: 数据库文件路径（覆盖配置文件）
- `TZ`: 时区设置（默认: Asia/Shanghai）

### 数据库结构

- `web_pages`: 网页列表
- `keywords`: 关键词列表
- `monitor_config`: 监控配置
- `subscribe_config`: 订阅配置
- `announcements`: 公告信息
- `push_config`: 推送配置

## 部署方案

### macOS 部署

#### 方式1: 源码部署

**前置要求:**
- Go 1.22+
- Node.js 18+
- npm

**步骤:**

1. **克隆项目**
```bash
git clone <repository-url>
cd demo-scrapy
```

2. **配置环境**
```bash
# 创建配置文件
cp config.yaml.example config.yaml
# 编辑配置文件，设置邮件等信息
vim config.yaml
```

3. **启动服务**
```bash
# 方式1: 使用启动脚本（推荐）
./start.sh

# 方式2: 手动启动
# 终端1: 启动后端
make build
./bin/server

# 终端2: 启动前端
cd frontend
npm install
npm run serve
```

4. **访问服务**
- 前端: http://localhost:5001
- 后端 API: http://localhost:5080
- Swagger: http://localhost:5080/swagger/index.html

#### 方式2: Docker 部署

```bash
# 构建镜像
make docker-build

# 启动服务
make docker-run

# 访问
# 前端: http://localhost
# 后端: http://localhost:5080
```

### Windows 部署

#### 方式1: 源码部署

**前置要求:**
- Go 1.22+
- Node.js 18+
- Git for Windows

**步骤:**

1. **克隆项目**
```cmd
git clone <repository-url>
cd demo-scrapy
```

2. **配置环境**
```cmd
copy config.yaml.example config.yaml
notepad config.yaml
```

3. **启动服务**
```cmd
REM 方式1: 使用批处理脚本
start.bat

REM 方式2: 手动启动
REM 终端1: 启动后端
go build -o bin\server.exe cmd\server\main.go
bin\server.exe

REM 终端2: 启动前端
cd frontend
npm install
npm run serve
```

#### 方式2: Docker Desktop 部署

```cmd
docker-compose up -d
```

### Linux 部署

#### 方式1: 源码部署

**前置要求:**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y golang-go nodejs npm git

# CentOS/RHEL
sudo yum install -y golang nodejs npm git
```

**步骤:**

1. **克隆项目**
```bash
git clone <repository-url>
cd demo-scrapy
```

2. **配置环境**
```bash
cp config.yaml.example config.yaml
vim config.yaml
```

3. **构建和启动**
```bash
# 构建后端
make build

# 启动后端（后台运行）
nohup ./bin/server > server.log 2>&1 &

# 构建前端
cd frontend
npm install
npm run build

# 使用 Nginx 部署前端（可选）
sudo cp -r dist/* /var/www/html/
```

4. **使用 systemd 管理服务（可选）**

创建 `/etc/systemd/system/demo-scrapy.service`:
```ini
[Unit]
Description=Demo Scrapy Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/demo-scrapy
ExecStart=/path/to/demo-scrapy/bin/server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
sudo systemctl daemon-reload
sudo systemctl enable demo-scrapy
sudo systemctl start demo-scrapy
```

#### 方式2: Docker 部署

```bash
# 构建镜像
make docker-build

# 启动服务
make docker-run

# 查看日志
make docker-logs
```

#### 方式3: 使用部署脚本

```bash
# Go 版本部署
./deploy.sh

# Node.js 版本部署（需要 backend-node 目录）
./deploy-node.sh
```

### Docker 部署

#### 快速开始

```bash
# 1. 构建镜像
make docker-build

# 2. 启动服务
make docker-run

# 3. 查看状态
docker-compose ps

# 4. 查看日志
make docker-logs

# 5. 停止服务
make docker-stop
```

#### 使用国内镜像源

如果网络较慢，可以使用国内镜像源:

```bash
# 使用国内镜像源构建
make docker-build-cn

# 使用国内镜像源启动
docker-compose -f docker-compose.cn.yml up -d
```

#### Docker Compose 配置说明

`docker-compose.yml` 包含两个服务:

- **backend**: Go 后端服务（端口 5080）
- **frontend**: Vue 前端服务（端口 80）

前端自动代理 `/api` 和 `/swagger` 请求到后端。

#### 数据持久化

数据库文件存储在 `./data` 目录，确保该目录有写权限:

```bash
mkdir -p data
chmod 755 data
```

## 开发指南

### 本地开发

**后端开发:**

```bash
# 安装依赖
go mod download

# 运行开发服务器
make run-dev

# 或使用启动脚本
./start.sh
```

**前端开发:**

```bash
cd frontend
npm install
npm run serve
```

### 构建生产版本

**后端构建:**

```bash
make build
# 生成 bin/server
```

**前端构建:**

```bash
cd frontend
npm run build
# 生成 dist/ 目录
```

### 代码结构

```
demo-scrapy/
├── cmd/server/          # 主程序入口
├── internal/            # 内部包
│   ├── api/            # API 路由和处理
│   ├── config/         # 配置管理
│   ├── crawler/        # 爬虫模块
│   ├── database/       # 数据库操作
│   ├── email/          # 邮件发送
│   ├── models/         # 数据模型
│   └── scheduler/      # 定时任务
├── frontend/           # 前端代码
│   ├── src/           # 源代码
│   ├── public/        # 静态资源
│   └── dist/         # 构建输出
├── docs/              # 文档
├── config.yaml        # 配置文件
└── Dockerfile         # Docker 配置
```

## API 文档

### Swagger 文档

启动服务后访问: http://localhost:5080/swagger/index.html

### 主要 API 端点

- `GET /api/web-pages` - 获取网页列表
- `POST /api/web-pages` - 创建网页
- `PUT /api/web-pages/:id` - 更新网页
- `DELETE /api/web-pages/:id` - 删除网页

- `GET /api/keywords` - 获取关键词列表
- `POST /api/keywords` - 创建关键词
- `DELETE /api/keywords/:id` - 删除关键词

- `GET /api/monitor-config` - 获取监控配置
- `POST /api/monitor-config` - 创建监控配置
- `PUT /api/monitor-config/:id` - 更新监控配置
- `DELETE /api/monitor-config/:id` - 删除监控配置

- `GET /api/subscribe-config` - 获取订阅配置
- `POST /api/subscribe-config` - 创建订阅配置
- `PUT /api/subscribe-config/:id` - 更新订阅配置
- `DELETE /api/subscribe-config/:id` - 删除订阅配置

- `GET /api/announcements` - 获取公告列表（支持分页和筛选）
- `GET /api/push-config` - 获取推送配置
- `PUT /api/push-config` - 更新推送配置

## 常见问题

### 1. 端口被占用

```bash
# macOS/Linux
lsof -ti:5080 | xargs kill -9

# Windows
netstat -ano | findstr :5080
taskkill /PID <PID> /F
```

### 2. 数据库锁定

删除数据库锁定文件:
```bash
rm -f monitor.db-shm monitor.db-wal
```

### 3. 邮件发送失败

检查 SMTP 配置:
- 确保 `smtp_pass` 是授权码（不是登录密码）
- QQ 邮箱需要开启 SMTP 服务并获取授权码
- 检查防火墙设置

### 4. 内存不足

使用编译后的二进制文件而不是 `go run`:
```bash
make build
./bin/server
```

### 5. Docker 构建失败

使用国内镜像源:
```bash
make docker-build-cn
```

## 许可证

[添加许可证信息]

## 联系方式

[添加联系方式]

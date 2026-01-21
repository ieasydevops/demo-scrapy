# Windows 环境安装说明

## 系统要求

- Windows 7 或更高版本（推荐 Windows 10/11）
- 64位操作系统
- 至少 100MB 可用磁盘空间

## 安装步骤

### 1. 下载安装包

下载 `server-windows-amd64.exe` 和 `config.yaml` 文件到同一目录。

### 2. 配置系统

#### 2.1 编辑配置文件

使用文本编辑器（如记事本）打开 `config.yaml` 文件，根据实际需求修改以下配置：

```yaml
web_pages:
  - name: "深圳政府采购网"
    url: "http://zfcg.szggzy.com:8081/gsgg/secondPage.html"

keywords:
  - "生态环境局"

monitor_configs:
  - web_page_name: "深圳政府采购网"
    crawl_time: "9"        # 采集时间（小时，0-23）
    crawl_freq: "daily"    # 采集频率：daily（每天）、hourly（每小时）
    keywords:
      - "生态环境局"

email:
  smtp_host: "smtp.qq.com"
  smtp_user: "your_email@qq.com"
  smtp_pass: "your_smtp_password"

server:
  port: 8080
  db_path: "./monitor.db"
```

#### 2.2 配置说明

- **web_pages**: 采集的目标网站列表
  - `name`: 网站名称
  - `url`: 网站URL地址

- **keywords**: 全局关键词列表，用于过滤公告

- **monitor_configs**: 监控配置列表
  - `web_page_name`: 对应的网站名称（需与 web_pages 中的 name 一致）
  - `crawl_time`: 定时采集时间（小时，0-23）
  - `crawl_freq`: 采集频率
    - `daily`: 每天执行一次
    - `hourly`: 每小时执行一次
  - `keywords`: 该监控任务的关键词列表

- **email**: 邮件通知配置
  - `smtp_host`: SMTP服务器地址
  - `smtp_user`: 发件人邮箱
  - `smtp_pass`: SMTP授权码（不是邮箱密码）

- **server**: 服务器配置
  - `port`: API服务端口（默认8080）
  - `db_path`: 数据库文件路径

### 3. 运行服务

#### 方式一：命令行运行

1. 打开命令提示符（CMD）或 PowerShell
2. 切换到程序所在目录
3. 执行以下命令：

```cmd
server.exe
```

#### 方式二：作为 Windows 服务运行（推荐）

1. 下载并安装 [NSSM](https://nssm.cc/download)（Non-Sucking Service Manager）
2. 以管理员身份打开命令提示符
3. 切换到 NSSM 安装目录
4. 执行以下命令安装服务：

```cmd
nssm install ProcurementMonitor "C:\path\to\server.exe"
nssm set ProcurementMonitor AppDirectory "C:\path\to\your\app\directory"
nssm start ProcurementMonitor
```

#### 方式三：开机自启动

1. 按 `Win + R` 打开运行对话框
2. 输入 `shell:startup` 并回车
3. 在打开的文件夹中创建快捷方式，指向 `server.exe`

### 4. 验证安装

服务启动后，可以通过以下方式验证：

1. **访问 API 文档**：打开浏览器访问 `http://localhost:8080/swagger/index.html`
2. **查看日志**：程序会在控制台输出运行日志
3. **检查数据库**：程序会在当前目录生成 `monitor.db` 数据库文件

### 5. 前端部署（可选）
如果需要使用 Web 界面：

1. 安装 Node.js（推荐 18 或更高版本）
2. 进入 `frontend` 目录
3. 执行以下命令：

```cmd
npm install
npm run build
```

4. 将 `frontend/dist` 目录中的文件部署到 Web 服务器（如 Nginx、IIS）

## 常见问题

### Q: 程序无法启动，提示端口被占用

A: 修改 `config.yaml` 中的 `server.port` 为其他端口（如 8081）

### Q: 如何查看程序运行日志？

A: 如果通过命令行运行，日志会直接显示在控制台。如果作为服务运行，可以通过 NSSM 查看日志输出

### Q: 如何停止服务？

A: 
- 命令行运行：按 `Ctrl + C`
- 服务运行：在服务管理器中停止服务，或使用命令 `nssm stop ProcurementMonitor`

### Q: 配置文件修改后需要重启吗？

A: 是的，修改配置文件后需要重启服务才能生效

### Q: 数据库文件在哪里？

A: 数据库文件位置由 `config.yaml` 中的 `server.db_path` 配置决定，默认为当前目录下的 `monitor.db`

## 技术支持

如遇到问题，请检查：
1. 配置文件格式是否正确（YAML格式）
2. 网络连接是否正常
3. 防火墙是否允许程序访问网络
4. 端口是否被其他程序占用

#!/bin/bash

if [ ! -f .env ]; then
    echo "创建.env文件..."
    cat > .env << EOF
SMTP_USER=403608355@qq.com
SMTP_PASS=your_smtp_password
SMTP_HOST=smtp.qq.com
EOF
    echo "请编辑.env文件，设置SMTP密码"
fi

mkdir -p data

echo "检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go，请先安装 Go 1.21 或更高版本"
    exit 1
fi

echo "✓ Go 版本: $(go version | awk '{print $3}')"

echo ""
echo "安装依赖..."
go mod download

echo ""
echo "生成 Swagger 文档..."
if command -v swag &> /dev/null; then
    swag init -g cmd/server/main.go -o docs
    echo "✓ Swagger 文档已生成"
elif go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs 2>/dev/null; then
    echo "✓ Swagger 文档已生成"
else
    echo "⚠ Swagger 文档生成失败，但可以继续运行"
fi

echo ""
echo "启动后端服务..."
echo "后端API: http://localhost:8080"
echo "Swagger UI: http://localhost:8080/swagger/index.html"
echo ""

export DB_PATH=./monitor.db
go run cmd/server/main.go

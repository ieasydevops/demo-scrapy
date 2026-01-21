#!/bin/bash

set -e

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

echo "检查依赖..."

if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go，请先安装 Go 1.21 或更高版本"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "错误: 未找到 Node.js，请先安装 Node.js 18 或更高版本"
    exit 1
fi

echo "✓ Go 版本: $(go version | awk '{print $3}')"
echo "✓ Node.js 版本: $(node --version)"

echo ""
echo "安装后端依赖..."
go mod download

echo ""
echo "生成 Swagger 文档..."
if command -v swag &> /dev/null; then
    swag init -g cmd/server/main.go -o docs 2>/dev/null && echo "✓ Swagger 文档已生成" || echo "⚠ Swagger 文档生成失败"
elif go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs 2>/dev/null; then
    echo "✓ Swagger 文档已生成"
else
    echo "⚠ Swagger 文档生成失败，但可以继续运行"
fi

echo ""
echo "安装前端依赖..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "正在安装前端依赖..."
    npm install
else
    echo "前端依赖已安装"
fi
cd ..

echo ""
echo "启动服务..."
echo "后端API: http://localhost:8080"
echo "Swagger UI: http://localhost:8080/swagger/index.html"
echo "前端: http://localhost:3000"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo ""

cleanup() {
    echo ""
    echo "正在停止服务..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

export DB_PATH=./monitor.db

go run cmd/server/main.go &
BACKEND_PID=$!

sleep 2

cd frontend
npm run serve > /dev/null 2>&1 &
FRONTEND_PID=$!
cd ..

wait

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

echo "启动服务..."
echo "后端API: http://localhost:5080"
echo "Swagger UI: http://localhost:5080/swagger/index.html"
echo "前端: http://localhost:5001"
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

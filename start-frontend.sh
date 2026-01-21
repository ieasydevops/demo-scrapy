#!/bin/bash

echo "检查 Node.js 环境..."
if ! command -v node &> /dev/null; then
    echo "错误: 未找到 Node.js，请先安装 Node.js 18 或更高版本"
    exit 1
fi

echo "✓ Node.js 版本: $(node --version)"

cd frontend

echo ""
echo "安装前端依赖..."
if [ ! -d "node_modules" ]; then
    npm install
else
    echo "前端依赖已安装"
fi

echo ""
echo "启动前端服务..."
echo "前端: http://localhost:5001"
echo ""

npm run serve

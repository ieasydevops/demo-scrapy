#!/bin/bash

set -e

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

VERSION=${1:-"1.0.0"}
BUILD_DIR="build"
DEPLOY_DIR="${BUILD_DIR}/demo-scrapy"

echo "开始构建版本: ${VERSION}"

rm -rf ${BUILD_DIR}
mkdir -p ${DEPLOY_DIR}

echo "构建前端..."
cd frontend
npm run build
cd ..
cp -r frontend/dist ${DEPLOY_DIR}/

echo "使用 Docker 构建后端 (linux/amd64)..."
docker run --rm \
  -v "$(pwd)":/app \
  -w /app \
  golang:1.24-alpine \
  sh -c "apk add --no-cache gcc musl-dev && go build -o build/demo-scrapy/server ./cmd/server"

echo "复制配置文件..."
cp config.yaml ${DEPLOY_DIR}/

cat > ${DEPLOY_DIR}/run.sh << 'EOF'
#!/bin/bash
export DB_PATH=./monitor.db
./server
EOF
chmod +x ${DEPLOY_DIR}/run.sh

echo "创建部署包..."
cd ${BUILD_DIR}
tar -czvf demo-scrapy-${VERSION}.tar.gz demo-scrapy
cd ..

echo "构建完成！"
echo "部署包: ${BUILD_DIR}/demo-scrapy-${VERSION}.tar.gz"
echo ""
echo "部署方式:"
echo "  1. 上传到目标机器"
echo "  2. tar -xzvf demo-scrapy-${VERSION}.tar.gz"
echo "  3. cd demo-scrapy && ./run.sh"

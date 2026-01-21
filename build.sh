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

echo "复制源码和配置..."
cp -r cmd ${DEPLOY_DIR}/
cp -r internal ${DEPLOY_DIR}/
cp -r docs ${DEPLOY_DIR}/
cp go.mod go.sum ${DEPLOY_DIR}/
cp config.yaml ${DEPLOY_DIR}/
cp -r frontend/dist ${DEPLOY_DIR}/

cat > ${DEPLOY_DIR}/build-and-run.sh << 'EOF'
#!/bin/bash
set -e
echo "下载依赖..."
go mod download
echo "编译..."
go build -o server ./cmd/server
echo "启动服务..."
export DB_PATH=./monitor.db
./server
EOF
chmod +x ${DEPLOY_DIR}/build-and-run.sh

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
echo "  3. cd demo-scrapy"
echo "  4. ./build-and-run.sh  # 首次运行（编译+启动）"
echo "  5. ./run.sh            # 后续运行"

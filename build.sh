#!/bin/bash

set -e

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

VERSION=${1:-"1.0.0"}
BUILD_DIR="build"
BIN_DIR="${BUILD_DIR}/bin"

echo "开始构建版本: ${VERSION}"

rm -rf ${BUILD_DIR}
mkdir -p ${BIN_DIR}

echo "构建后端 (linux/amd64)..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o ${BIN_DIR}/server ./cmd/server

echo "构建前端..."
cd frontend
npm run build
cd ..
cp -r frontend/dist ${BUILD_DIR}/

echo "复制配置文件..."
cp config.yaml ${BUILD_DIR}/

echo "构建完成！"
echo "后端: ${BIN_DIR}/server"
echo "前端: ${BUILD_DIR}/dist"

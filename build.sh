#!/bin/bash

set -e

VERSION=${1:-"1.0.0"}
BUILD_DIR="build"
BIN_DIR="${BUILD_DIR}/bin"

echo "开始构建版本: ${VERSION}"

rm -rf ${BUILD_DIR}
mkdir -p ${BIN_DIR}

echo "构建 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o ${BIN_DIR}/server-linux-amd64 ./cmd/server

echo "构建 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o ${BIN_DIR}/server-windows-amd64.exe ./cmd/server

echo "构建 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o ${BIN_DIR}/server-darwin-amd64 ./cmd/server
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o ${BIN_DIR}/server-darwin-arm64 ./cmd/server

echo "复制配置文件..."
cp config.yaml ${BUILD_DIR}/
cp -r frontend ${BUILD_DIR}/ 2>/dev/null || true

echo "创建 Windows 安装包..."
mkdir -p ${BUILD_DIR}/windows
cp ${BIN_DIR}/server-windows-amd64.exe ${BUILD_DIR}/windows/server.exe
cp config.yaml ${BUILD_DIR}/windows/
cp README-Windows.md ${BUILD_DIR}/windows/ 2>/dev/null || true

echo "构建完成！"
echo "Linux 版本: ${BIN_DIR}/server-linux-amd64"
echo "Windows 版本: ${BIN_DIR}/server-windows-amd64.exe"
echo "macOS 版本: ${BIN_DIR}/server-darwin-amd64, ${BIN_DIR}/server-darwin-arm64"

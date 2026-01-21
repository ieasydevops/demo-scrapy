#!/bin/bash

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="$PROJECT_DIR/build/deploy"
BIN_DIR="$PROJECT_DIR/bin"
VERSION=$(date +%Y%m%d%H%M%S)

REMOTE_HOST="${DEPLOY_HOST:-122.114.197.58}"
REMOTE_USER="${DEPLOY_USER:-xinyun.mei}"
REMOTE_PASS="${DEPLOY_PASS:-mima@haha123}"
REMOTE_DIR="${DEPLOY_DIR:-/home/xinyun.mei/code/github.com/ieasydevops/deploy}"

PACKAGE_NAME="demo-scrapy-${VERSION}.tar.gz"
PID_FILE="server.pid"

print_usage() {
    echo "用法: $0 {pack|deploy|start|stop|restart|status|clean|all}"
    echo ""
    echo "命令说明:"
    echo "  pack     - 一键打包: 编译程序并打包部署文件"
    echo "  deploy   - 一键部署: 部署到远程服务器"
    echo "  start    - 一键启动: 启动服务"
    echo "  stop     - 停止服务"
    echo "  restart  - 重启服务"
    echo "  status   - 查看服务状态"
    echo "  clean    - 一键清理: 清理构建文件和临时文件"
    echo "  all      - 执行全部: pack -> deploy -> start"
    echo ""
    echo "环境变量:"
    echo "  DEPLOY_HOST  - 远程服务器地址 (默认: $REMOTE_HOST)"
    echo "  DEPLOY_USER  - 远程服务器用户 (默认: $REMOTE_USER)"
    echo "  DEPLOY_PASS  - 远程服务器密码 (默认: 已设置)"
    echo "  DEPLOY_DIR   - 远程部署目录 (默认: $REMOTE_DIR)"
}

pack() {
    echo "========== [1/4] 开始打包 v${VERSION} =========="
    
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    mkdir -p "$BIN_DIR"
    
    echo "[1/4-1] 编译Go程序（支持跨平台）..."
    cd "$PROJECT_DIR"
    
    TARGET_OS="${BUILD_OS:-linux}"
    TARGET_ARCH="${BUILD_ARCH:-amd64}"
    
    echo "目标平台: ${TARGET_OS}/${TARGET_ARCH}"
    
    if [ "$TARGET_OS" = "windows" ]; then
        BINARY_NAME="server.exe"
    else
        BINARY_NAME="server"
    fi
    
    CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build \
        -ldflags="-s -w" \
        -o "$BUILD_DIR/$BINARY_NAME" \
        ./cmd/server/main.go
    
    if [ ! -f "$BUILD_DIR/server" ]; then
        echo "错误: 编译失败"
        exit 1
    fi
    
    echo "[1/4-2] 复制配置文件..."
    cp -f "$PROJECT_DIR/config.yaml" "$BUILD_DIR/"
    
    echo "[1/4-3] 创建启动脚本..."
    
    if [ "$TARGET_OS" = "windows" ]; then
        cat > "$BUILD_DIR/run.bat" << 'EOF'
@echo off
cd /d "%~dp0"

set PID_FILE=server.pid
set LOG_FILE=server.log

:start
if exist "%PID_FILE%" (
    for /f "tokens=*" %%a in (%PID_FILE%) do set PID=%%a
    tasklist /FI "PID eq %PID%" 2>nul | find /I /N "%PID%">nul
    if "%ERRORLEVEL%"=="0" (
        echo 服务已在运行 (PID: %PID%)
        goto :end
    )
)

start /B server.exe -config config.yaml > "%LOG_FILE%" 2>&1
echo %ERRORLEVEL% > %PID_FILE%
timeout /t 1 /nobreak >nul
tasklist /FI "IMAGENAME eq server.exe" 2>nul | find /I /N "server.exe">nul
if "%ERRORLEVEL%"=="0" (
    echo 服务启动成功
) else (
    echo 服务启动失败，查看日志: type %LOG_FILE%
    exit /b 1
)
goto :end

:stop
if exist "%PID_FILE%" (
    for /f "tokens=*" %%a in (%PID_FILE%) do set PID=%%a
    taskkill /PID %PID% /F >nul 2>&1
    del /f /q %PID_FILE% >nul 2>&1
    echo 服务已停止
) else (
    echo 服务未运行
)
goto :end

:restart
call :stop
timeout /t 1 /nobreak >nul
call :start
goto :end

:status
tasklist /FI "IMAGENAME eq server.exe" 2>nul | find /I /N "server.exe">nul
if "%ERRORLEVEL%"=="0" (
    echo 服务运行中
) else (
    echo 服务未运行
)
goto :end

:end
EOF
    else
        cat > "$BUILD_DIR/run.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"

PID_FILE="server.pid"
LOG_FILE="server.log"

start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            echo "服务已在运行 (PID: $PID)"
            return
        fi
    fi
    
    nohup ./server -config config.yaml > "$LOG_FILE" 2>&1 &
    echo $! > $PID_FILE
    echo "服务已启动 (PID: $(cat $PID_FILE))"
    sleep 1
    if kill -0 $(cat $PID_FILE) 2>/dev/null; then
        echo "服务启动成功"
    else
        echo "服务启动失败，查看日志: tail -f $LOG_FILE"
        exit 1
    fi
}

stop() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            kill $PID
            rm -f $PID_FILE
            echo "服务已停止"
            return
        fi
    fi
    echo "服务未运行"
}

restart() {
    stop
    sleep 1
    start
}

status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            echo "服务运行中 (PID: $PID)"
            return
        fi
    fi
    echo "服务未运行"
}

case "$1" in
    start)   start ;;
    stop)    stop ;;
    restart) restart ;;
    status)  status ;;
    *)       echo "用法: $0 {start|stop|restart|status}"; exit 1 ;;
esac
EOF
        chmod +x "$BUILD_DIR/run.sh"
    fi
    
    echo "[1/4-4] 创建部署包..."
    cd "$BUILD_DIR"
    tar -czf "$PROJECT_DIR/$PACKAGE_NAME" .
    
    echo "打包完成: $PACKAGE_NAME"
    echo "文件大小: $(du -h "$PROJECT_DIR/$PACKAGE_NAME" | cut -f1)"
}

deploy() {
    echo "========== [2/4] 开始部署到远程服务器 =========="
    
    if [ ! -f "$PROJECT_DIR/$PACKAGE_NAME" ]; then
        echo "错误: 部署包不存在，请先执行打包: $0 pack"
        exit 1
    fi
    
    echo "[2/4-1] 检查远程服务器连接..."
    if ! command -v sshpass &> /dev/null; then
        echo "错误: 未安装 sshpass，请安装:"
        echo "  macOS: brew install hudochenkov/sshpass/sshpass"
        echo "  Ubuntu: sudo apt-get install sshpass"
        echo "  CentOS: sudo yum install sshpass"
        exit 1
    fi
    
    echo "[2/4-2] 上传部署包到远程服务器..."
    sshpass -p "$REMOTE_PASS" scp -o StrictHostKeyChecking=no \
        "$PROJECT_DIR/$PACKAGE_NAME" \
        "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/"
    
    echo "[2/4-3] 解压部署包..."
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && \
         mkdir -p backup && \
         [ -f server ] && mv server backup/server.\$(date +%Y%m%d%H%M%S) || true && \
         tar -xzf $PACKAGE_NAME && \
         chmod +x server run.sh && \
         rm -f $PACKAGE_NAME"
    
    echo "[2/4-4] 验证部署文件..."
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && \
         [ -f server ] && [ -f config.yaml ] && [ -f run.sh ] && \
         echo '部署文件验证成功' || echo '警告: 部分文件缺失'"
    
    echo "部署完成"
}

start() {
    echo "========== [3/4] 启动服务 =========="
    
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && ./run.sh start"
    
    echo "服务启动完成"
}

stop() {
    echo "========== 停止服务 =========="
    
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && ./run.sh stop"
    
    echo "服务已停止"
}

restart() {
    echo "========== 重启服务 =========="
    
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && ./run.sh restart"
    
    echo "服务重启完成"
}

status() {
    echo "========== 服务状态 =========="
    
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \
        "$REMOTE_USER@$REMOTE_HOST" \
        "cd $REMOTE_DIR && ./run.sh status"
}

clean() {
    echo "========== [4/4] 清理构建文件 =========="
    
    echo "[4/4-1] 清理本地构建目录..."
    rm -rf "$BUILD_DIR"
    rm -rf "$BIN_DIR"
    
    echo "[4/4-2] 清理部署包..."
    rm -f "$PROJECT_DIR"/demo-scrapy-*.tar.gz
    
    echo "[4/4-3] 清理Go缓存..."
    go clean -cache -modcache 2>/dev/null || true
    
    echo "清理完成"
}

all() {
    echo "========== 执行完整流程 =========="
    pack
    deploy
    start
    echo ""
    echo "========== 部署完成 =========="
    echo "版本: $VERSION"
    echo "远程目录: $REMOTE_DIR"
    echo ""
    echo "查看日志: ssh $REMOTE_USER@$REMOTE_HOST 'tail -f $REMOTE_DIR/server.log'"
    echo "查看状态: $0 status"
}

CMD="${1:-}"
if [ "$CMD" = "pack" ] && [ -n "${2:-}" ]; then
    VERSION="$2"
fi

case "$CMD" in
    pack)
        pack
        ;;
    deploy)
        if [ -n "${2:-}" ]; then
            VERSION="$2"
        fi
        deploy
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    clean)
        clean
        ;;
    all)
        if [ -n "${2:-}" ]; then
            VERSION="$2"
        fi
        all
        ;;
    *)
        print_usage
        exit 1
        ;;
esac

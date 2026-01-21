#!/bin/bash
set -e

VERSION=${1:-1.0.0}
REMOTE_HOST="122.114.197.58"
REMOTE_USER="xinyun.mei"
REMOTE_PASS="mima@haha123"
REMOTE_DIR="/home/xinyun.mei/code/github.com/ieasydevops/deploy"

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="$PROJECT_DIR/build/deploy"

echo "========== 开始部署 v${VERSION} =========="

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

echo "[1/5] 构建前端..."
cd "$PROJECT_DIR/frontend"
npm install --silent
npm run build
cp -r dist "$BUILD_DIR/frontend"

echo "[2/5] 打包源码和依赖..."
cd "$PROJECT_DIR"
go mod vendor
cp -r cmd internal docs go.mod go.sum config.yaml vendor "$BUILD_DIR/"

cat > "$BUILD_DIR/run.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"

PID_FILE="server.pid"

start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            echo "服务已在运行 (PID: $PID)"
            return
        fi
    fi
    nohup ./server -config config.yaml > server.log 2>&1 &
    echo $! > $PID_FILE
    echo "服务已启动 (PID: $!)"
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
    *)       echo "Usage: $0 {start|stop|restart|status}" ;;
esac
EOF
chmod +x "$BUILD_DIR/run.sh"

echo "[3/5] 打包并上传到服务器..."
tar -czf "$PROJECT_DIR/build/deploy.tar.gz" -C "$BUILD_DIR" .

if ! command -v sshpass &> /dev/null; then
    echo "警告: 未安装sshpass，请手动输入密码"
    SSH_CMD="ssh"
    SCP_CMD="scp"
else
    SSH_CMD="sshpass -p '$REMOTE_PASS' ssh"
    SCP_CMD="sshpass -p '$REMOTE_PASS' scp"
fi

eval $SSH_CMD -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "mkdir -p $REMOTE_DIR"
eval $SCP_CMD -o StrictHostKeyChecking=no "$PROJECT_DIR/build/deploy.tar.gz" $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/
eval $SSH_CMD -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && rm -rf cmd internal docs go.mod go.sum config.yaml vendor frontend run.sh server server.pid server.log 2>/dev/null; tar -xzf deploy.tar.gz && rm deploy.tar.gz"

echo "[4/5] 在服务器上构建..."
eval $SSH_CMD -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && export PATH=\$PATH:/usr/local/go/bin && export GOPROXY=direct && export GOSUMDB=off && go clean -modcache 2>/dev/null || true && go build -mod=vendor -ldflags '-X main.version=${VERSION}' -o server ./cmd/server && ls -la server"

echo "[5/5] 启动服务..."
eval $SSH_CMD -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && chmod +x server run.sh && ./run.sh restart"

echo "========== 部署完成 =========="
echo "后端API: http://$REMOTE_HOST:5080"
echo "Swagger: http://$REMOTE_HOST:5080/swagger/index.html"

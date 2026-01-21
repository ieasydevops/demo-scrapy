#!/bin/bash

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

cd frontend

echo "启动前端服务..."
echo "前端: http://localhost:5001"
echo ""

npm run serve

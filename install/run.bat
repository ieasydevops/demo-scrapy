@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

set PID_FILE=server.pid

if "%1"=="start" goto start
if "%1"=="stop" goto stop
if "%1"=="restart" goto restart
if "%1"=="status" goto status
echo Usage: %~nx0 {start|stop|restart|status}
exit /b

:start
if exist "%PID_FILE%" (
    for /f "tokens=*" %%a in (%PID_FILE%) do set PID=%%a
    tasklist /FI "PID eq %PID%" 2>nul | find /I /N "%PID%">nul
    if not errorlevel 1 (
        echo 服务已在运行 (PID: %PID%)
        exit /b
    )
)

set NODE_ENV=production
set DB_PATH=.\monitor.db

if not exist "backend-node\node_modules" (
    echo 安装依赖...
    cd backend-node
    call npm install --production --silent
    cd ..
)

start /b node backend-node\server.js > server.log 2>&1
timeout /t 1 /nobreak >nul
for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq node.exe" /FO LIST ^| find "PID"') do (
    echo %%a > %PID_FILE%
    echo 服务已启动 (PID: %%a)
    exit /b
)
echo 服务启动失败，请查看 server.log
exit /b

:stop
if exist "%PID_FILE%" (
    for /f "tokens=*" %%a in (%PID_FILE%) do set PID=%%a
    tasklist /FI "PID eq %PID%" 2>nul | find /I /N "%PID%">nul
    if not errorlevel 1 (
        taskkill /PID %PID% /F >nul 2>&1
        del /q "%PID_FILE%"
        echo 服务已停止
        exit /b
    )
)
echo 服务未运行
exit /b

:restart
call %~nx0 stop
timeout /t 2 /nobreak >nul
call %~nx0 start
exit /b

:status
if exist "%PID_FILE%" (
    for /f "tokens=*" %%a in (%PID_FILE%) do set PID=%%a
    tasklist /FI "PID eq %PID%" 2>nul | find /I /N "%PID%">nul
    if not errorlevel 1 (
        echo 服务运行中 (PID: %PID%)
        exit /b
    )
)
echo 服务未运行
exit /b

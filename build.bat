@echo off
setlocal enabledelayedexpansion

set VERSION=%1
if "%VERSION%"=="" set VERSION=1.0.0

set BUILD_DIR=build
set BIN_DIR=%BUILD_DIR%\bin

echo 开始构建版本: %VERSION%

if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
mkdir %BUILD_DIR%
mkdir %BIN_DIR%

echo 构建 Windows 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o %BIN_DIR%\server-windows-amd64.exe cmd\server\main.go

echo 构建 Linux 版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o %BIN_DIR%\server-linux-amd64.exe cmd\server\main.go

echo 复制配置文件...
copy config.yaml %BUILD_DIR%\

echo 创建 Windows 安装包...
mkdir %BUILD_DIR%\windows
copy %BIN_DIR%\server-windows-amd64.exe %BUILD_DIR%\windows\server.exe
copy config.yaml %BUILD_DIR%\windows\
copy README-Windows.md %BUILD_DIR%\windows\ 2>nul

echo 构建完成！
echo Windows 版本: %BIN_DIR%\server-windows-amd64.exe
echo Linux 版本: %BIN_DIR%\server-linux-amd64.exe

endlocal

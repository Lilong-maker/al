@echo off
setlocal

REM 设置Go环境
set "GOTOOLCHAIN=local"
set "GOROOT=C:\Program Files\Go"
set "PATH=C:\Program Files\Go\bin;%PATH%"

REM 切换到工作目录
cd /d "C:\Users\Lenovo\Desktop\al\srv\dasic\cmd"

REM 运行程序
echo [1/2] 正在编译...
"C:\Program Files\Go\bin\go.exe" build -o temp_app.exe main.go
if errorlevel 1 (
    echo [错误] 编译失败！
    pause
    exit /b 1
)

echo [2/2] 正在运行程序...
temp_app.exe

REM 清理
del temp_app.exe 2>nul

echo.
echo [完成] 程序已退出
pause

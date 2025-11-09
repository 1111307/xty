@echo off
chcp 65001 > nul
echo === Nginx安全快速核查 ===
echo.

echo 1. 检查Nginx进程...
tasklist /fi "imagename eq nginx.exe" /fo table

echo.
echo 2. 检查配置文件中的用户设置...
findstr "user" conf\nginx.conf >nul && (echo  找到用户配置： && findstr "user" conf\nginx.conf) || echo  未配置运行用户

echo.
echo 3. 检查SSL配置...
findstr "listen.*443" conf\nginx.conf >nul && echo  已配置HTTPS || echo  未配置HTTPS
findstr "ssl_certificate" conf\nginx.conf >nul && echo  找到SSL证书配置 || echo  未找到SSL证书配置

echo.
echo 4. 检查版本隐藏...
findstr "server_tokens off" conf\nginx.conf >nul && echo  已隐藏版本号 || echo  未隐藏版本号

echo.
echo 5. 检查监听的端口...
netstat -ano | findstr ":80 :443"

echo.
echo === 核查完成 ===
pause
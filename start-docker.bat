@echo off
REM Docker 快速启动脚本 - Windows

echo ========================================
echo   供应商名片管理系统 - Docker 启动
echo ========================================
echo.

echo 构建并启动服务...
docker-compose up -d --build

echo.
echo ========================================
echo   部署完成！
echo ========================================
echo   访问地址: http://localhost:8080
echo   健康检查: http://localhost:8080/health
echo ========================================
echo.
echo 查看日志: docker-compose logs -f
echo 停止服务: docker-compose down
echo.
pause

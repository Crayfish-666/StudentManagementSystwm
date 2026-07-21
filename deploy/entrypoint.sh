#!/bin/sh
# =============================================================================
# StudentHub 容器启动脚本 (Java Spring Boot 版)
#   职责: 准备运行时目录 -> 打印启动信息 -> exec java -jar
# 设计原则:
#   1. 失败立即退出 (set -e)
#   2. 不吞错, 关键错误打印到 stderr 便于 docker logs 检索
#   3. exec 替换 shell 进程, 让 JVM 成为 PID 1, 正确接收 SIGTERM
# =============================================================================

set -e

# 运行时目录 (与 Dockerfile 中创建的目录保持一致)
DATA_DIR="/app/data"
STORAGE_DIR="/app/storage"
LOG_DIR="/app/logs"

# 确保运行时目录存在 (首次启动时挂载的命名卷是空目录)
mkdir -p "$DATA_DIR" "$STORAGE_DIR" "$LOG_DIR"

# 启动横幅
echo "=================================================="
echo "  StudentHub Container Starting (Java Spring Boot)"
echo "  Env       : ${APP_ENV:-dev}"
echo "  Port      : ${SERVER_PORT:-8080}"
echo "  Data dir  : $DATA_DIR"
echo "  Storage   : $STORAGE_DIR"
echo "  Logs      : $LOG_DIR"
echo "  Timezone  : ${TZ:-Asia/Shanghai}"
echo "  JVM Opts  : ${JAVA_OPTS:--Xms512m -Xmx1024m}"
echo "=================================================="

# 关键环境变量校验 (避免生产使用默认值)
if [ "$APP_ENV" = "prod" ] || [ "$APP_ENV" = "production" ]; then
    if [ -z "$APP_DATA_KEY" ]; then
        echo "[FATAL] 生产环境必须设置 APP_DATA_KEY 环境变量 (AES 字段级加密密钥, 32 字节)" >&2
        exit 1
    fi
fi

# exec 替换当前 shell 进程, 让 JVM 作为 PID 1
# 优势: 正确接收 docker stop 的 SIGTERM 信号, 实现优雅关闭
# JAVA_OPTS 用于 JVM 调优 (内存/GC 等), 默认 -Xms512m -Xmx1024m
exec java ${JAVA_OPTS:--Xms512m -Xmx1024m} \
    -Duser.timezone=${TZ:-Asia/Shanghai} \
    -jar /app/app.jar

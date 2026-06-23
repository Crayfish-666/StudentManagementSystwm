#!/bin/sh
# =============================================================================
# StudentHub 容器启动脚本
#   职责: 准备运行时目录 -> 打印启动信息 -> exec 后端二进制
# 设计原则:
#   1. 失败立即退出 (set -e)
#   2. 不吞错, 关键错误打印到 stderr 便于 docker logs 检索
#   3. exec 替换 shell 进程, 让后端成为 PID 1, 正确接收 SIGTERM
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
echo "  StudentHub Container Starting"
echo "  Env       : ${APP_ENV:-dev}"
echo "  Port      : 8080"
echo "  Data dir  : $DATA_DIR"
echo "  Storage   : $STORAGE_DIR"
echo "  Logs      : $LOG_DIR"
echo "  Timezone  : ${TZ:-Asia/Shanghai}"
echo "=================================================="

# 关键环境变量校验 (避免生产使用默认值)
if [ "$APP_ENV" = "prod" ] || [ "$APP_ENV" = "production" ]; then
    if [ -z "$JWT_SECRET" ]; then
        echo "[FATAL] 生产环境必须设置 JWT_SECRET 环境变量" >&2
        exit 1
    fi
    # 注: cryptox 包读取 CRYPTOX_KEY (非 ADR 文档中描述的 APP_DATA_KEY)
    # 如需字段级加密 (身份证/银行卡), 请同步设置 CRYPTOX_KEY
    if [ -z "$CRYPTOX_KEY" ]; then
        echo "[WARN] CRYPTOX_KEY 未设置, 字段级加密 (AES) 将不可用" >&2
    fi
fi

# exec 替换当前 shell 进程, 让后端作为 PID 1
# 优势: 正确接收 docker stop 的 SIGTERM 信号, 实现优雅关闭
exec /app/server

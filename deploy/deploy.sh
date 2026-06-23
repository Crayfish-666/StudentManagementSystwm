#!/bin/sh
# =============================================================================
# StudentHub 服务器端部署脚本
# -----------------------------------------------------------------------------
# 职责: 拉取新镜像 -> 替换运行 -> 健康检查 -> 失败回滚
# 调用方: GitHub Actions (.github/workflows/ci.yml / release.yml)
# 必需环境变量:
#   REGISTRY_URL       镜像仓库地址 (如 harbor.example.com)
#   REGISTRY_USERNAME  镜像仓库用户
#   REGISTRY_PASSWORD  镜像仓库密码
#   IMAGE_TAG          本次发布的镜像 tag (如 v1.0.0 / develop-<sha>)
# 可选环境变量:
#   DEPLOY_ENV         dev | production (默认 dev, 影响镜像 tag 后缀)
#   DEPLOY_PATH        项目部署目录 (默认 /opt/studenthub)
#   HEALTH_CHECK_URL   健康检查 URL (默认 http://127.0.0.1:8080/api/v1/healthz)
#   HEALTH_TIMEOUT     健康检查超时秒数 (默认 90)
# =============================================================================

set -e

# -------- 1. 参数校验 --------
: "${REGISTRY_URL:?[FATAL] REGISTRY_URL 未设置}"
: "${REGISTRY_USERNAME:?[FATAL] REGISTRY_USERNAME 未设置}"
: "${REGISTRY_PASSWORD:?[FATAL] REGISTRY_PASSWORD 未设置}"
: "${IMAGE_TAG:?[FATAL] IMAGE_TAG 未设置}"

DEPLOY_ENV="${DEPLOY_ENV:-dev}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/studenthub}"
HEALTH_CHECK_URL="${HEALTH_CHECK_URL:-http://127.0.0.1:8080/api/v1/healthz}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-90}"
IMAGE_NAME="${IMAGE_NAME:-studenthub}"
FULL_IMAGE="${REGISTRY_URL}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "=================================================="
echo "  StudentHub Deploy"
echo "  Env       : ${DEPLOY_ENV}"
echo "  Path      : ${DEPLOY_PATH}"
echo "  Image     : ${FULL_IMAGE}"
echo "  Health    : ${HEALTH_CHECK_URL}"
echo "=================================================="

cd "${DEPLOY_PATH}"

# -------- 2. 登录镜像仓库 --------
echo "[1/5] 登录镜像仓库 ${REGISTRY_URL} ..."
echo "${REGISTRY_PASSWORD}" | docker login "${REGISTRY_URL}" \
    -u "${REGISTRY_USERNAME}" --password-stdin

# -------- 3. 拉取新镜像 --------
echo "[2/5] 拉取新镜像 ${FULL_IMAGE} ..."
docker pull "${FULL_IMAGE}"

# -------- 4. 记录上一个 image tag (用于回滚) --------
PREVIOUS_TAG=""
if [ -f ./.deploy_state ]; then
    PREVIOUS_TAG=$(cat ./.deploy_state)
    echo "    上一个版本: ${PREVIOUS_TAG}"
fi

# -------- 5. 切换镜像并滚动重启 --------
echo "[3/5] 切换镜像并滚动重启 ..."
# 导出 IMAGE_TAG 给 docker compose 使用
export IMAGE_TAG
# docker compose v2: docker compose (空格), v1: docker-compose (横线)
if docker compose version >/dev/null 2>&1; then
    docker compose pull
    docker compose up -d --remove-orphans
else
    docker-compose pull
    docker-compose up -d --remove-orphans
fi

# -------- 6. 健康检查 --------
echo "[4/5] 健康检查 (timeout=${HEALTH_TIMEOUT}s) ..."
WAITED=0
HEALTH_OK=0
while [ "${WAITED}" -lt "${HEALTH_TIMEOUT}" ]; do
    if wget -qO- "${HEALTH_CHECK_URL}" >/dev/null 2>&1; then
        HEALTH_OK=1
        break
    fi
    sleep 3
    WAITED=$((WAITED + 3))
    echo "    等待中... ${WAITED}s"
done

# -------- 7. 结果处理 --------
echo "[5/5] 处理结果 ..."
if [ "${HEALTH_OK}" -eq 1 ]; then
    echo "${IMAGE_TAG}" > ./.deploy_state
    # 清理悬虚镜像
    docker image prune -f >/dev/null 2>&1 || true
    echo "=================================================="
    echo "  部署成功: ${FULL_IMAGE}"
    echo "=================================================="
    exit 0
fi

# 健康检查失败 -> 回滚
echo "[FATAL] 健康检查失败, 启动回滚流程 ..."
if [ -z "${PREVIOUS_TAG}" ]; then
    echo "[FATAL] 无可回滚版本, 请人工介入" >&2
    exit 1
fi

export IMAGE_TAG="${PREVIOUS_TAG}"
if docker compose version >/dev/null 2>&1; then
    docker compose up -d --remove-orphans
else
    docker-compose up -d --remove-orphans
fi
sleep 5
if wget -qO- "${HEALTH_CHECK_URL}" >/dev/null 2>&1; then
    echo "=================================================="
    echo "  回滚成功: ${PREVIOUS_TAG}"
    echo "=================================================="
    exit 1
fi

echo "[FATAL] 回滚后仍无法恢复, 请人工介入" >&2
exit 2

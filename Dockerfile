# =============================================================================
# StudentHub · 多阶段构建（Java Spring Boot 版）
#   阶段 1: 构建前端 (Node 20 + pnpm)
#   阶段 2: 编译后端 (Maven + JDK 21)
#   阶段 3: 运行时镜像 (Eclipse Temurin JRE 21) - 集成后端 jar + 前端 dist
# 镜像内布局:
#   /app/app.jar                   后端 Fat JAR
#   /app/frontend/dist/            前端静态资源
#   /app/data/                     SQLite + WAL/SHM (持久化卷)
#   /app/storage/                  上传文件 (持久化卷)
#   /app/logs/                     日志输出 (持久化卷)
#   /app/deploy/entrypoint.sh      启动脚本
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1 · 前端构建
# -----------------------------------------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# 单独拷贝 lock 文件以充分利用 Docker 缓存
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# 安装 pnpm (使用 npm 全局安装, 避开 corepack 在国内网络下的兼容问题)
RUN npm config set registry https://registry.npmmirror.com \
    && npm install -g pnpm@9.15.4 \
    && pnpm config set registry https://registry.npmmirror.com \
    && pnpm install --no-frozen-lockfile

# 拷贝源码并构建
COPY frontend/ ./
RUN pnpm run build

# -----------------------------------------------------------------------------
# Stage 2 · 后端编译 (Maven + JDK 21)
# -----------------------------------------------------------------------------
FROM maven:3.9-eclipse-temurin-21 AS backend-builder

WORKDIR /build/backend

# 配置阿里云 Maven 镜像加速国内构建
RUN mkdir -p /root/.m2 && cat > /root/.m2/settings.xml <<'EOF'
<settings>
  <mirrors>
    <mirror>
      <id>aliyun</id>
      <mirrorOf>central</mirrorOf>
      <url>https://maven.aliyun.com/repository/public</url>
    </mirror>
  </mirrors>
</settings>
EOF

# 单独拷贝 pom.xml 以利用 Docker 缓存下载依赖
COPY backend/pom.xml ./
RUN mvn -B dependency:go-offline -q

# 拷贝源码并编译打包（跳过测试，测试在 CI 阶段运行）
COPY backend/src ./src
RUN mvn -B clean package -DskipTests -q \
    && mv target/studenthub-backend-*.jar /out/app.jar

# -----------------------------------------------------------------------------
# Stage 3 · 运行时镜像
# -----------------------------------------------------------------------------
FROM eclipse-temurin:21-jre-alpine

# 安装运行时基础依赖: 时区数据 / ca-certificates / wget (健康检查)
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache \
        ca-certificates \
        tzdata \
        wget \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户运行服务 (安全基线)
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 拷贝后端 Fat JAR
COPY --from=backend-builder /out/app.jar /app/app.jar

# 拷贝前端构建产物
COPY --from=frontend-builder /build/frontend/dist /app/frontend/dist

# 拷贝启动脚本并去除 Windows 换行符 (\r\n -> \n)
COPY deploy/entrypoint.sh /app/deploy/entrypoint.sh
RUN sed -i 's/\r$//' /app/deploy/entrypoint.sh \
    && chmod +x /app/deploy/entrypoint.sh

# 创建运行时数据目录并授权
RUN mkdir -p /app/data /app/storage /app/logs \
    && chown -R app:app /app

USER app

# Spring Boot 默认监听 :8080 (application.yml 中配置)
EXPOSE 8080

# 容器内健康检查 (基于 Spring Boot Actuator /actuator/health)
HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/v1/actuator/health >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/deploy/entrypoint.sh"]

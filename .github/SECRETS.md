# StudentHub · GitHub Actions 凭据清单

> 本文档列出 CI/CD 流水线所需的全部 GitHub **Secrets** 和 **Variables**。
> 在 `Settings → Secrets and variables → Actions` 中配置。

---

## 一、Variables (变量, 公开可读)

| 名称 | 必填 | 默认 | 说明 |
| ---- | ---- | ---- | ---- |
| `IMAGE_NAME` | 否 | `studenthub` | 镜像名称, 与 `deploy/.env.example` 的 `IMAGE_NAME` 保持一致 |

---

## 二、Secrets (凭据, 加密存储)

### 2.1 镜像仓库

| 名称 | 必填 | 示例 | 说明 |
| ---- | ---- | ---- | ---- |
| `REGISTRY_URL` | 是 | `harbor.example.com` | 镜像仓库地址, **不含协议前缀**; 留空则 CI 仅构建不推送 |
| `REGISTRY_USERNAME` | 是 | `robot$studenthub` | 仓库用户, 推荐使用机器人账号 |
| `REGISTRY_PASSWORD` | 是 | `xxxxx` | 仓库密码或 token |

> Harbor / GitLab Registry / 阿里云 ACR / 腾讯云 TCR / DockerHub 均可, 只需 `docker login` 能通过。

### 2.2 SSH 部署

| 名称 | 必填 | 示例 | 说明 |
| ---- | ---- | ---- | ---- |
| `SSH_HOST` | 是 | `203.0.113.10` | 部署目标主机 IP / 域名 |
| `SSH_PORT` | 否 | `22` | SSH 端口, 默认 22 |
| `SSH_USERNAME` | 是 | `deploy` | 部署用户, 推荐非 root + `docker` 用户组 |
| `SSH_PRIVATE_KEY` | 是 | `-----BEGIN OPENSSH PRIVATE KEY-----...` | SSH 私钥 (推荐用 ed25519, 与公钥配对) |
| `DEPLOY_PATH` | 否 | `/opt/studenthub` | 服务器项目目录, 必须包含 `docker-compose.yml` 和 `deploy/deploy.sh` |

### 2.3 生产环境保护

`release.yml` 中的 `deploy` 任务配置了 `environment: production`, 需在 `Settings → Environments → production` 中:
- 配置 `Required reviewers` (审批人), 实现生产发布的"双人复核"
- 可选: 配置部署分支限制 (仅允许 `main` / `v*.*.*`)

---

## 三、配置示例

### 3.1 Harbor (自建)

```
REGISTRY_URL    = harbor.example.com
REGISTRY_USERNAME = robot$studenthub-ci
REGISTRY_PASSWORD = <机器人账号 token>
```

### 3.2 阿里云 ACR

```
REGISTRY_URL    = registry.cn-hangzhou.aliyuncs.com
REGISTRY_USERNAME = <阿里云账号 / RAM 用户>
REGISTRY_PASSWORD = <固定密码 / 临时 token>
```

### 3.3 SSH 部署用户准备 (服务器端一次性操作)

```bash
# 1. 创建部署用户并加入 docker 组
sudo useradd -m -s /bin/bash deploy
sudo usermod -aG docker deploy

# 2. 准备部署目录
sudo mkdir -p /opt/studenthub/{data,storage,logs}
sudo chown -R deploy:deploy /opt/studenthub
cd /opt/studenthub
# 3. 拷贝 docker-compose.yml / deploy/deploy.sh / .env
#    (注意 deploy.sh 的最终执行者是 deploy 用户)

# 4. 将 CI 公钥加入 authorized_keys
mkdir -p ~/.ssh
echo "ssh-ed25519 AAAA... github-actions-deploy" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

---

## 四、首次启用步骤

1. **配置 Secrets**: 按上表在 GitHub 仓库设置中添加
2. **配置 Variables** (可选): 添加 `IMAGE_NAME`
3. **配置 Environments**: 创建 `dev` 和 `production` 两个环境
4. **触发 PR 检查**: 任意 PR 打开后, `pr-check.yml` 自动运行
5. **触发 dev 部署**: 推送代码到 `develop` 分支, `ci.yml` 自动构建 + 部署
6. **触发 prod 部署**: 推送 `v*.*.*` 标签 (如 `git tag v1.0.0 && git push origin v1.0.0`), `release.yml` 自动构建 + 发布

---

## 五、调试

- **查看 workflow 运行历史**: 仓库 `Actions` Tab
- **查看镜像推送日志**: 展开对应 Job → `Build and push` 步骤
- **查看部署日志**: 展开 `Trigger remote deploy` 步骤, ssh-action 会原样回显脚本输出
- **手动触发**: 任一 workflow 均配置了 `workflow_dispatch`, 可在 Actions 页面手动运行

---

## 六、ADR 引用

- §3.4 代码风格与质量门禁
- §3.5 分支与发布策略
- §3.9 测试规范
- §5 部署与运维

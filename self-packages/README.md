# 自用预编译包

生产机内存不够时，不要在服务器上执行 `pnpm build` 或 `go build -tags embed`。
用这里打好的 linux 包直接替换二进制。

当前版本：`0.2.4-self.2`

包含：

- 重置余额兑换码 + 7 天提醒 + 管理员「执行重置」
- 管理员可手选下次自然重置日
- 只读管理员（可看账号状态，不能改数据，也不能导出 OAuth / API Key）
- 嵌入管理后台（`-tags embed`），单文件即可提供前端

## 下载

| 架构 | 文件 |
| --- | --- |
| x86_64 / amd64 | [sub2api_0.2.4-self.2_linux_amd64.tar.gz](./sub2api_0.2.4-self.2_linux_amd64.tar.gz) |
| arm64 / aarch64 | [sub2api_0.2.4-self.2_linux_arm64.tar.gz](./sub2api_0.2.4-self.2_linux_arm64.tar.gz) |
| 校验 | [checksums.txt](./checksums.txt) |

不要用官方 `install.sh upgrade` 或后台「检测更新」。那会拉取 `Wei-Shaw/sub2api`，把自用功能盖掉。

## 已有 systemd 安装（`/opt/sub2api`）

把本仓库里的脚本拷到服务器后：

```bash
sudo VERSION=0.2.4-self.2 bash deploy/upgrade-self.sh
```

或手动：

```bash
arch=$(uname -m)
case "$arch" in x86_64|amd64) a=amd64 ;; aarch64|arm64) a=arm64 ;; esac
ver=0.2.4-self.2
curl -fL -o /tmp/sub2api.tgz \
  "https://github.com/wanjunping97-cyber/sub2api/raw/cursor/self-package-current-fec0/self-packages/sub2api_${ver}_linux_${a}.tar.gz"
sudo tar -xzf /tmp/sub2api.tgz -C /tmp sub2api
sudo systemctl stop sub2api
sudo cp /opt/sub2api/sub2api /opt/sub2api/sub2api.bak
sudo install -m 0755 /tmp/sub2api /opt/sub2api/sub2api
sudo systemctl start sub2api
/opt/sub2api/sub2api --version
```

启动后会自动跑数据库迁移。

## 还没有安装过

不要在服务器上编译。可以二选一：

1. 先用官方 `install.sh` 只装一次 systemd / 目录骨架，然后立刻按上面的步骤换成这个预编译包。之后不要再跑官方 `upgrade`。
2. 自己准备 PostgreSQL、Redis，把包里的 `sub2api` 放到 `/opt/sub2api/sub2api`，用包内 `deploy/sub2api.service` 启动。首次运行会走安装向导，或设置 `AUTO_SETUP=true` 后用环境变量自动初始化。

## Docker

不要 `docker pull weishaw/sub2api:latest`。把上面的 `sub2api` 二进制挂进容器，或在内存够的机器上按仓库 Dockerfile 构建自己的镜像后再拷到服务器。

## 以后自己打包

在一台内存充足的机器上：

```bash
VERSION=0.2.4-self.2 bash backend/scripts/package-self.sh
```

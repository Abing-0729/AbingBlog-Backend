# AbingBlog 部署与调试手册

适用于一台 Ubuntu 22.04/24.04 阿里云 ECS。前后端同机部署：宿主机 Nginx 对外监听 80/443，后端占用 `127.0.0.1:8080`，前端占用 `127.0.0.1:8081`。

## 首次准备

阿里云安全组只开放 TCP `22`、`80`、`443`，不要开放 `3306`、`8080`、`8081`。

```bash
ssh root@SERVER_IP
apt update
apt install -y ca-certificates curl git nginx openssl
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker nginx
mkdir -p /opt/abingblog-backend /opt/abingblog-frontend
```

将两个仓库的 `docker-compose.prod.yml` 分别上传到对应目录。

## GitHub Actions 配置

两个仓库共用：

```text
DEPLOY_HOST
DEPLOY_USER
DEPLOY_SSH_KEY
```

后端额外配置：

```text
DEPLOY_PATH=/opt/abingblog-backend
MYSQL_ROOT_PASSWORD=强密码
JWT_SECRET=至少32位随机字符串
```

前端额外配置：

```text
FRONTEND_DEPLOY_PATH=/opt/abingblog-frontend
```

前后端同域名时不要设置 `VITE_API_BASE_URL`。如果使用独立 API 域名，在前端 production Environment Variables 设置 `VITE_API_BASE_URL`。

GitHub Actions 的 Workflow permissions 需要允许 `Read and write permissions`，以便推送 GHCR 镜像。

## 发布

先发布后端，再发布前端。两个仓库分别执行：

```bash
git tag v1.0.0
git push origin v1.0.0
```

流水线会构建镜像、推送 GHCR、SSH 部署并执行健康检查。

```bash
docker compose -f /opt/abingblog-backend/docker-compose.prod.yml ps
docker compose -f /opt/abingblog-frontend/docker-compose.prod.yml ps
docker compose -f /opt/abingblog-backend/docker-compose.prod.yml logs --tail=200 server
docker compose -f /opt/abingblog-frontend/docker-compose.prod.yml logs --tail=200 frontend
```

## 初始化管理员

后端镜像包含 seed 程序。首次部署后执行：

```bash
cd /opt/abingblog-backend
docker compose -f docker-compose.prod.yml run --rm server /app/seed
```

默认账号是 `admin` / `admin123`。生产环境建议设置 `ABINGBLOG_ADMIN_PASSWORD` 后再次执行 seed。用户已存在时会跳过创建。

不要执行 `docker compose down -v`，否则会删除 MySQL 数据卷。

## 宿主机 Nginx

创建 `/etc/nginx/sites-available/abingblog`：

```nginx
server {
    listen 80;
    server_name example.com www.example.com;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

启用配置：

```bash
ln -s /etc/nginx/sites-available/abingblog /etc/nginx/sites-enabled/abingblog
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx
```

## HTTPS

先把域名 A 记录指向 ECS 公网 IP，并确认 HTTP 可访问：

```bash
curl -I http://example.com
apt install -y certbot python3-certbot-nginx
certbot --nginx -d example.com -d www.example.com
certbot renew --dry-run
```

Certbot 会申请免费的 Let's Encrypt 证书、修改 Nginx 并配置续期。

## 健康检查与调试

```bash
curl http://127.0.0.1:8080/api/v1/healthz
curl http://127.0.0.1:8081/healthz
curl -I http://example.com
docker ps
docker stats
df -h
free -h
```

后端连不上数据库时检查 `MYSQL_HOST=mysql`、MySQL healthcheck 和容器日志。页面能打开但 API 失败时检查 Nginx `/api/` 代理和后端 8080 健康检查。登录失败时确认已执行 seed，且 `JWT_SECRET` 长度至少 32 个字符。

## 回滚

流水线镜像按 commit SHA 保存。回滚时在对应目录执行：

```bash
export IMAGE=ghcr.io/OWNER/REPO:OLD_SHA
docker compose -f docker-compose.prod.yml up -d
```

确认正常后再执行 `docker image prune -f`。

## 备份

```bash
docker exec $(docker ps -qf name=mysql) mysqldump -uroot -p'PASSWORD' blog > /opt/abingblog-backup-$(date +%F).sql
```

恢复前停止后端，恢复后重新启动。密码不要写入 Git 或公开日志。

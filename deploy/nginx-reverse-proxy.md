# Nginx 反向代理部署指南

本文档介绍如何将 GitHub Pages 博客（yiiewang.github.io）通过 Nginx 反向代理部署到自有服务器，并配置 HTTPS。

## 📋 目录

- [前置条件](#前置条件)
- [部署步骤](#部署步骤)
  - [1. 服务器准备](#1-服务器准备)
  - [2. 创建配置文件](#2-创建配置文件)
  - [3. 启动 HTTP 服务](#3-启动-http-服务)
  - [4. 配置域名解析](#4-配置域名解析)
  - [5. 申请 SSL 证书](#5-申请-ssl-证书)
  - [6. 启用 HTTPS](#6-启用-https)
  - [7. 配置证书自动续期](#7-配置证书自动续期)
- [常用命令](#常用命令)
- [故障排查](#故障排查)

---

## 前置条件

- 一台 Linux 服务器（推荐 CentOS/Ubuntu/Debian）
- 已安装 Docker
- 一个已解析到服务器的域名
- 服务器防火墙已开放 80 和 443 端口

---

## 部署步骤

### 1. 服务器准备

#### 1.1 安装 Docker（如未安装）

```bash
# CentOS/RHEL
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
```

#### 1.2 安装 Certbot（用于 SSL 证书）

```bash
# CentOS/RHEL 8+
dnf install -y certbot

# CentOS/RHEL 7
yum install -y epel-release
yum install -y certbot

# Ubuntu/Debian
apt update
apt install -y certbot
```

#### 1.3 开放防火墙端口

```bash
# 腾讯云/阿里云等云服务器需要在控制台安全组开放端口

# 如果使用 firewalld
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --permanent --add-port=443/tcp
firewall-cmd --reload

# 如果使用 ufw
ufw allow 80/tcp
ufw allow 443/tcp
```

### 2. 创建配置文件

```bash
# 创建目录
mkdir -p /opt/nginx-proxy
```

创建 HTTP 版本配置（用于首次部署，申请证书前）：

```bash
cat > /opt/nginx-proxy/nginx.conf << 'EOF'
# HTTP 版本（用于首次部署申请 SSL 证书前）
worker_processes auto;
events { worker_connections 1024; }

http {
    resolver 8.8.8.8 valid=300s;
    
    server {
        listen 80;
        server_name cloaks.cn www.cloaks.cn;
        
        location / {
            proxy_pass https://yiiewang.github.io;
            proxy_ssl_server_name on;
            proxy_set_header Host yiiewang.github.io;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
}
EOF
```

### 3. 启动 HTTP 服务

```bash
# 启动 Nginx 容器（HTTP 模式）
docker run -d \
  --name nginx-proxy \
  --restart unless-stopped \
  -p 80:80 \
  -v /opt/nginx-proxy/nginx.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine

# 验证是否启动成功
docker ps | grep nginx-proxy
curl -I http://localhost
```

### 4. 配置域名解析

在域名服务商控制台添加 A 记录：

| 主机记录 | 记录类型 | 记录值 |
|---------|---------|--------|
| @ | A | 你的服务器IP |
| www | A | 你的服务器IP |

等待 DNS 生效（通常 1-10 分钟），验证：

```bash
ping cloaks.cn
curl -I http://cloaks.cn
```

### 5. 申请 SSL 证书

```bash
# 停止 Nginx 释放 80 端口（certbot 需要）
docker stop nginx-proxy

# 申请证书（替换为你的域名和邮箱）
certbot certonly --standalone \
  -d cloaks.cn \
  -d www.cloaks.cn \
  --agree-tos \
  --email your-email@example.com

# 如果不想提供邮箱
certbot certonly --standalone \
  -d cloaks.cn \
  -d www.cloaks.cn \
  --agree-tos \
  --register-unsafely-without-email
```

证书申请成功后，会显示：

```
Certificate is saved at: /etc/letsencrypt/live/cloaks.cn/fullchain.pem
Key is saved at: /etc/letsencrypt/live/cloaks.cn/privkey.pem
```

### 6. 启用 HTTPS

更新配置文件为 HTTPS 版本：

```bash
cat > /opt/nginx-proxy/nginx.conf << 'EOF'
worker_processes auto;
events { worker_connections 1024; }

http {
    resolver 8.8.8.8 valid=300s;
    
    # HTTP -> HTTPS 重定向
    server {
        listen 80;
        server_name cloaks.cn www.cloaks.cn;
        return 301 https://$host$request_uri;
    }
    
    # HTTPS 反向代理
    server {
        listen 443 ssl http2;
        server_name cloaks.cn www.cloaks.cn;
        
        # SSL 证书路径（Let's Encrypt）
        ssl_certificate /etc/letsencrypt/live/cloaks.cn/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/cloaks.cn/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
        ssl_prefer_server_ciphers off;
        
        location / {
            # 代理到 GitHub Pages
            proxy_pass https://yiiewang.github.io;
            proxy_ssl_server_name on;
            proxy_set_header Host yiiewang.github.io;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # 缓存设置
            proxy_cache_valid 200 10m;
            proxy_buffering on;
        }
    }
}
EOF
```

重新部署容器：

```bash
# 删除旧容器
docker rm -f nginx-proxy

# 启动新容器（挂载证书）
docker run -d \
  --name nginx-proxy \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -v /opt/nginx-proxy/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v /etc/letsencrypt:/etc/letsencrypt:ro \
  nginx:alpine

# 验证 HTTPS
curl -I https://cloaks.cn
curl -I https://www.cloaks.cn
```

### 7. 配置证书自动续期

Let's Encrypt 证书有效期 90 天，需要定期续期。

```bash
# 创建续期脚本
cat > /opt/nginx-proxy/renew-ssl.sh << 'EOF'
#!/bin/bash
# SSL 证书续期脚本
# 建议添加到 crontab: 0 3 1 * * /opt/nginx-proxy/renew-ssl.sh

set -e

DOMAIN="cloaks.cn"
CONTAINER_NAME="nginx-proxy"

echo "[$(date)] 开始续期 SSL 证书..."

# 停止 nginx 释放 80 端口
docker stop $CONTAINER_NAME

# 续期证书
certbot renew --standalone

# 重启 nginx
docker start $CONTAINER_NAME

echo "[$(date)] SSL 证书续期完成！"
EOF

# 添加执行权限
chmod +x /opt/nginx-proxy/renew-ssl.sh

# 添加定时任务（每月 1 号凌晨 3 点执行）
(crontab -l 2>/dev/null; echo "0 3 1 * * /opt/nginx-proxy/renew-ssl.sh >> /var/log/ssl-renew.log 2>&1") | crontab -

# 验证定时任务
crontab -l
```

---

## 常用命令

### 容器管理

```bash
# 查看容器状态
docker ps | grep nginx-proxy

# 查看日志
docker logs nginx-proxy
docker logs -f nginx-proxy  # 实时查看

# 重启容器（修改配置后）
docker restart nginx-proxy

# 停止/启动容器
docker stop nginx-proxy
docker start nginx-proxy

# 进入容器
docker exec -it nginx-proxy sh

# 测试配置是否正确
docker exec nginx-proxy nginx -t

# 重新加载配置（不重启容器）
docker exec nginx-proxy nginx -s reload
```

### SSL 证书管理

```bash
# 查看证书信息
certbot certificates

# 查看证书有效期
openssl x509 -in /etc/letsencrypt/live/cloaks.cn/cert.pem -noout -dates

# 手动续期
docker stop nginx-proxy
certbot renew
docker start nginx-proxy
```

---

## 故障排查

### 1. 网站无法访问

```bash
# 检查容器是否运行
docker ps | grep nginx-proxy

# 检查端口是否监听
netstat -tlnp | grep -E '80|443'

# 检查防火墙
firewall-cmd --list-ports  # firewalld
ufw status                  # ufw
```

### 2. SSL 证书问题

```bash
# 检查证书文件是否存在
ls -la /etc/letsencrypt/live/cloaks.cn/

# 测试 SSL 连接
openssl s_client -connect cloaks.cn:443 -servername cloaks.cn
```

### 3. 配置错误

```bash
# 测试 Nginx 配置
docker exec nginx-proxy nginx -t

# 查看详细错误日志
docker logs nginx-proxy 2>&1 | tail -50
```

---

## 自定义配置

如需代理其他 GitHub Pages，修改以下内容：

| 配置项 | 说明 |
|--------|------|
| `server_name` | 你的域名 |
| `proxy_pass` | 目标 GitHub Pages 地址 |
| `proxy_set_header Host` | 目标 GitHub Pages 地址 |
| `ssl_certificate` | SSL 证书路径（对应你的域名） |
| `ssl_certificate_key` | SSL 私钥路径（对应你的域名） |

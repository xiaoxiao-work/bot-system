FROM alpine:3.18

# 安装 SSL 证书（用于 HTTPS 请求）
RUN apk add --no-cache ca-certificates

# 创建工作目录
WORKDIR /app

# 复制二进制文件
COPY ./bot-system /app/

# 设置可执行权限
RUN chmod +x /app/bot-system

# 暴露服务端口
EXPOSE 10006

# 启动应用
CMD ["/app/bot-system"]

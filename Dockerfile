# 纯运行镜像（本地预编译）
FROM alpine:latest

# 使用国内镜像并安装运行时依赖
ENV TZ=Asia/Shanghai
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates wget tzdata && \
    addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 复制预编译的二进制文件（从 dist/ 目录）
COPY build/supplier-card-manager ./main
RUN chmod +x ./main

# 创建必要的目录
RUN mkdir -p /app/data/uploads /app/data/images /app/data && \
    chown -R appuser:appuser /app

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
    CMD wget -qO- http://localhost:8080/health > /dev/null 2>&1 || exit 1

CMD ["./main"]

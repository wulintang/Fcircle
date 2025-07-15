# Build阶段：构建Go应用
FROM golang:1.20-alpine AS builder

WORKDIR /app
# 将项目中的文件复制到/app目录
COPY . .
RUN go mod download  # 下载Go依赖
RUN CGO_ENABLED=0 GOOS=linux go build -o /fcircle ./cmd/fetch  # 编译Go应用

# 运行阶段：构建最小的运行时镜像
FROM alpine:3.17

WORKDIR /app

# 从builder阶段复制构建好的二进制文件
COPY --from=builder /fcircle /app/

# 复制配置文件
COPY ./config/ /app/config/

# 将 start.sh 脚本复制到容器中
COPY ./start.sh /app/start.sh

# 给脚本赋予执行权限
RUN chmod +x /app/start.sh

# 设置容器启动时运行的命令
CMD ["/app/start.sh"]


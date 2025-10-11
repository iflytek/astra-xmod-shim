# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖（仅构建阶段需要）
RUN apk add --no-cache git gcc musl-dev

# 复制模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建静态二进制（禁用 CGO）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -o astron-xmod-shim ./cmd/server/main.go


# 第二阶段：运行阶段
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 安装 CA 证书（用于 HTTPS 请求）
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 从构建阶段复制**正确的二进制文件**
COPY --from=builder /app/astron-xmod-shim ./

# 注意：不要 COPY conf/，因为 Helm 会通过 ConfigMap 挂载配置！
# 如果你保留 COPY conf/，会导致 ConfigMap 挂载被覆盖（因为挂载会覆盖目录内容）
# 所以建议删除下面这行 👇
# COPY conf/ ./conf/

# 创建日志目录（可选，应用也可自行创建）
RUN mkdir -p /app/logs

# 暴露端口（与 values.yaml 中 service.port 一致，如 7777）
EXPOSE 7777

# 启动命令
CMD ["./astron-xmod-shim"]
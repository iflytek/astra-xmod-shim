# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

# 复制go.mod和go.sum文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o xmod-shim cmd/server/main.go

# 第二阶段：运行阶段
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/xmod-shim .

# 复制配置文件
COPY conf/ ./conf/

# 创建日志目录
RUN mkdir -p /app/logs

# 设置环境变量（可选）
ENV TZ=Asia/Shanghai

# 暴露应用端口（根据实际应用配置调整）
EXPOSE 8080

# 设置启动命令
CMD ["./astra-xmod-shim"]
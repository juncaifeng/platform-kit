FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/demo-service .

# 使用平台基础镜像
FROM platform-kit:latest

WORKDIR /app

COPY --from=builder /app/demo-service .

# 服务配置
ENV SERVICE_NAME=demo-service
ENV SERVICE_PORT=8080
ENV REGISTER_MODE=daemon
ENV REGISTRY_TYPE=consul
ENV REGISTRY_ADDR=http://consul:8500
ENV HEALTH_CHECK_PATH=/health

# 业务服务的启动命令（作为 entrypoint 的参数）
CMD ["./demo-service"]

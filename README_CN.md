# Kitex Shop

[English](./README.md) | 简体中文

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Kitex](https://img.shields.io/badge/Kitex-0.16.3-4C8BF5)](https://www.cloudwego.io/zh/docs/kitex/)
[![Hertz](https://img.shields.io/badge/Hertz-0.10.4-00A6A6)](https://www.cloudwego.io/zh/docs/hertz/)

Kitex Shop 是一个精简的 Go 微服务示例，用于演示基于 [CloudWeGo Hertz](https://www.cloudwego.io/zh/docs/hertz/) 和 [CloudWeGo Kitex](https://www.cloudwego.io/zh/docs/kitex/) 的 HTTP 到 RPC 完整调用链。仓库包含 Proto3 IDL、Kitex-Protobuf 生成代码、多阶段 Docker 镜像、Amazon ECR 发布脚本以及 AWS ECS Service Connect 任务定义。

项目已不再使用 etcd 服务发现。服务现在通过稳定的 DNS 名 `item` 和 `stock` 互相连接：本地容器运行时由 Docker 网络提供解析，部署到 AWS 时由 ECS Service Connect 提供解析。

## 系统架构

```text
客户端
  |
  | HTTP GET /api/item
  v
API（Hertz，:8889）
  |
  | Kitex-Protobuf RPC：item:8891
  v
Item 商品服务
  |
  | Kitex-Protobuf RPC：stock:8890
  v
Stock 库存服务
```

| 组件 | 技术 | 端口 | 职责 |
| --- | --- | ---: | --- |
| API | Hertz + Kitex 客户端 | `8889` | 暴露 HTTP 接口并调用 Item 服务 |
| Item 服务 | Kitex 服务端/客户端 | `8891` | 组装商品信息并请求库存数据 |
| Stock 服务 | Kitex 服务端 | `8890` | 返回指定商品的库存数据 |

当前示例接口固定请求商品 ID `1024`。Item 服务返回演示用商品信息，Stock 服务将商品 ID 原样作为库存值返回。该行为仅用于简化示例，背后没有数据库。

## 仓库结构

```text
.
├── api/                         # Hertz HTTP 网关
├── rpc/
│   ├── item/                    # Item Kitex 服务及 Stock 客户端
│   └── stock/                   # Stock Kitex 服务
├── idl/                         # Proto3 服务和消息定义
├── kitex_gen/                   # 根据 Proto3 IDL 生成的代码
├── generate.sh                 # 确定性的 Kitex 代码生成脚本
├── Dockerfile.api               # API 多阶段镜像
├── Dockerfile.item              # Item 服务多阶段镜像
├── Dockerfile.stock             # Stock 服务多阶段镜像
├── aws-ecr-*.sh                 # ECR 创建、登录、标记和推送脚本
├── aws-task-api.json            # API ECS Task Definition
├── aws-task-item.json           # Item ECS Task Definition
├── aws-task-stock.json          # Stock ECS Task Definition
├── aws-task.json                # 独立的 Service Connect Nginx 示例
├── AWS_ECS_SERVICE_CONNECT.md   # ECS Service Connect 部署说明
├── go.mod
└── go.sum
```

`kitex_gen/` 中的生成文件已提交到仓库，因此构建示例时不要求预先安装 Kitex 代码生成工具。

## 前置条件

构建或运行示例前，请安装：

- Go 1.22 或更高版本
- Git
- Docker，用于推荐的本地端到端运行方式

如需部署到 AWS，还需要：

- AWS CLI v2
- 已完成身份认证，并有权访问 ECR、ECS、IAM、CloudWatch Logs 和 Service Connect/Cloud Map 的 AWS 账号
- 在 Apple Silicon 上部署时，需要 Docker Buildx 或其他能够构建 Linux `amd64` 镜像的工具

## 从源码构建

克隆仓库并下载依赖：

```bash
git clone https://github.com/linzworld/kitex-shop.git
cd kitex-shop
go mod download
```

构建全部 Go 包：

```bash
go build ./...
```

应用源码使用 `item:8891` 和 `stock:8890` 作为服务地址。下面的 Docker 和 ECS 流程会自动解析这两个名称。如果直接在宿主机运行 Go 进程，需要自行提供等效的 DNS 或 hosts 映射。

## 使用 Docker 构建和运行

Docker 是最简单的本地端到端运行方式，因为容器名会提供与 ECS Service Connect 相同的 `item`、`stock` DNS 名。

构建三个镜像：

```bash
docker build -f Dockerfile.stock -t kitex-shop-stock:latest .
docker build -f Dockerfile.item -t kitex-shop-item:latest .
docker build -f Dockerfile.api -t kitex-shop-api:latest .
```

创建网络，并按照依赖顺序启动服务：

```bash
docker network create kitex-shop

docker run -d --name stock --network kitex-shop \
  -p 8890:8890 kitex-shop-stock:latest

docker run -d --name item --network kitex-shop \
  -p 8891:8891 kitex-shop-item:latest

docker run -d --name api --network kitex-shop \
  -p 8889:8889 kitex-shop-api:latest
```

验证完整调用链：

```bash
curl http://localhost:8889/api/item
```

响应是 Kitex 对商品对象的字符串表示，其中包含 ID `1024`、标题 `Kitex`、示例描述和库存值 `1024`。

如果服务没有正常启动，可以查看日志：

```bash
docker logs stock
docker logs item
docker logs api
```

使用完毕后删除本地容器和网络：

```bash
docker rm -f api item stock
docker network rm kitex-shop
```

## 测试

运行仓库级 Go 测试命令：

```bash
go test ./...
```

仓库包含 Handler 单元测试，覆盖正常响应、必填 ID 校验、Kitex 业务错误和库存降级行为。可使用上面的 Docker 启动流程和 `curl` 请求进行端到端冒烟测试。

## Protobuf 接口

服务契约位于 `idl/`：

- `base.proto` 定义通用的 `BaseResp` 消息。
- `item.proto` 定义 `Item`、`GetItemReq`、`GetItemResp` 和 `ItemService.GetItem`。
- `stock.proto` 定义库存请求、响应消息以及 `StockService.GetItemStock`。

服务使用基于 TTHeader/TCP 的 Kitex-Protobuf，而不是标准 gRPC。修改 IDL 后，请安装 Kitex v0.16.3 并重新生成全部提交到仓库的代码：

```bash
go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3
./generate.sh
go test ./...
```

除非明确需要维护生成代码补丁，否则不要手动编辑生成文件。

## 部署到 AWS ECS

面向部署的方案为每个应用服务准备独立镜像和 ECS Task Definition。

### 1. 构建 Linux 镜像

ECR 脚本使用下面的本地镜像标签。在 Apple Silicon 上，请显式构建 `linux/amd64` 镜像：

```bash
docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.stock -t kitex-shop-stock:latest .

docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.item -t kitex-shop-item:latest .

docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.api -t kitex-shop-api:latest .
```

### 2. 发布到 ECR

运行前请检查脚本中的 `AWS_REGION`、`AWS_ACCOUNT_ID` 和仓库名称：

```bash
./aws-ecr-kitex-shop-stock.sh
./aws-ecr-kitex-shop-item.sh
./aws-ecr-kitex-shop-api.sh
```

每个脚本都会验证 AWS 身份，按需创建 ECR 仓库，登录 Docker，标记并推送本地镜像，最后输出带标签和摘要的镜像 URI。

### 3. 注册 Task Definition

注册前请检查 JSON 文件中的账号 ID、Role ARN、区域、镜像 URI、CPU、内存和日志组：

```bash
aws ecs register-task-definition \
  --region us-west-1 \
  --cli-input-json file://aws-task-stock.json

aws ecs register-task-definition \
  --region us-west-1 \
  --cli-input-json file://aws-task-item.json

aws ecs register-task-definition \
  --region us-west-1 \
  --cli-input-json file://aws-task-api.json
```

### 4. 配置 Service Connect

按照以下顺序部署服务：

1. Stock 使用端口名称 `stock-rpc` 发布 `stock:8890`。
2. Item 使用端口名称 `item-rpc` 发布 `item:8891`，并连接 Stock。
3. API 以客户端身份加入命名空间，并连接 Item。

Kitex 使用基于 TTHeader/TCP 的原生 Protobuf 协议，因此 Item 和 Stock 的端口映射不能声明 `appProtocol: grpc`。API 端口可以声明 `appProtocol: http`。

命名空间、端口、日志、安全组、部署顺序、验证和排错的详细说明见 [AWS_ECS_SERVICE_CONNECT.md](./AWS_ECS_SERVICE_CONNECT.md)。

## 配置说明

- 服务地址目前直接写在应用中，分别为 `item:8891` 和 `stock:8890`。
- API 监听 `0.0.0.0:8889`，Item 监听 `0.0.0.0:8891`，Stock 监听 `0.0.0.0:8890`。
- AWS 配置文件包含特定环境的账号 ID、区域、ARN 和日志组名称。在其他 AWS 账号中使用前必须检查并替换。
- `aws-task.json` 是独立的 Nginx Service Connect 示例，不属于 Kitex Shop 调用链。
- 当前示例不再需要 etcd，也不包含 etcd 注册中心客户端。

## 当前限制

- HTTP 接口使用固定商品 ID，暂不接收请求参数。
- 响应使用 Kitex 字符串表示，而不是结构化 JSON API 契约。
- 暂无数据库、缓存、认证、重试策略和应用级可观测性。
- 暂无独立的单元测试和集成测试套件。
- 服务端点是硬编码值，尚未通过环境变量或配置文件注入。

这些限制使仓库能够集中展示 Hertz 到 Kitex 的调用链，以及 ECS Service Connect 部署模型。

## 相关文档

- [Kitex 文档](https://www.cloudwego.io/zh/docs/kitex/)
- [Kitex 快速开始](https://www.cloudwego.io/zh/docs/kitex/getting-started/tutorial/)
- [Hertz 文档](https://www.cloudwego.io/zh/docs/hertz/)
- [AWS ECS Service Connect 文档](https://docs.aws.amazon.com/zh_cn/AmazonECS/latest/developerguide/service-connect.html)

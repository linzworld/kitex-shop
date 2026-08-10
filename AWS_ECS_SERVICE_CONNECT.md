# AWS ECS Service Connect 规范

## 服务拓扑

```text
外部请求 → API:8889 → item:8891 → stock:8890
```

| 服务 | 端口 | Service Connect 角色 | 地址 |
| --- | ---: | --- | --- |
| API | 8889 | 客户端 | 不发布 |
| Item | 8891 | 客户端 + 服务端 | `item:8891` |
| Stock | 8890 | 服务端 | `stock:8890` |

统一配置：

```text
区域：us-west-1
命名空间：kitex-shop
ARN：arn:aws:servicediscovery:us-west-1:236763663116:namespace/ns-skt5xg2p35pzl2cg
```

## Task Definition

- family、容器名、镜像仓库使用统一服务名，如 `kitex-shop-stock`。
- 使用 `awsvpc` 和 `MANAGED_INSTANCES`。
- 服务端必须声明命名端口；`name` 必须与 Service Connect 的 `portName` 一致。
- API 使用 HTTP，可设置 `appProtocol: http`。
- Kitex/Thrift 使用普通 TCP，不设置 `appProtocol`，不能误设为 `grpc`。

Stock 端口示例：

```json
{
  "name": "stock-rpc",
  "containerPort": 8890,
  "hostPort": 8890,
  "protocol": "tcp"
}
```

任务定义修改后必须注册新 revision：

```bash
aws ecs register-task-definition \
  --region us-west-1 \
  --cli-input-json file://aws-task-stock.json
```

## Service Connect

Task Definition 的命名端口不会自动发布服务。选择“客户端和服务器”模式后，必须在 ECS Service 中添加端口配置。

Stock：

```json
{
  "enabled": true,
  "namespace": "arn:aws:servicediscovery:us-west-1:236763663116:namespace/ns-skt5xg2p35pzl2cg",
  "services": [{
    "portName": "stock-rpc",
    "discoveryName": "stock",
    "clientAliases": [{"dnsName": "stock", "port": 8890}]
  }]
}
```

Item 使用相同结构，将配置替换为：

```text
portName: item-rpc
discoveryName: item
dnsName: item
port: 8891
```

API 只作为客户端加入命名空间，不配置 `services`。

Managed Instances 必须使用 Capacity Provider Strategy；创建 Service 时不要指定 `launchType`。

## 日志

应用日志配置在 `containerDefinitions[].logConfiguration`：

```json
{
  "logDriver": "awslogs",
  "options": {
    "awslogs-group": "/ecs/kitex-shop-stock-service",
    "awslogs-region": "us-west-1",
    "awslogs-stream-prefix": "ecs"
  }
}
```

- 生产环境建议提前创建日志组并设置保留期。
- 使用 `awslogs-create-group: true` 时，Execution Role 需要 `logs:CreateLogGroup`。
- Execution Role 需要 ECR 拉取、`logs:CreateLogStream` 和 `logs:PutLogEvents` 权限。
- Service Connect 代理日志在 ECS Service 中单独配置，建议使用 `/ecs/kitex-shop-stock-service-connect`。

## 网络与部署

- Stock `8890` 仅允许 Item 安全组访问。
- Item `8891` 仅允许 API 安全组访问。
- API `8889` 仅允许 ALB 安全组访问。
- 共用安全组时，`8890`、`8891` 可使用安全组自身作为来源。
- 私有子网需要 NAT，或配置 ECR、CloudWatch Logs 等 VPC Endpoint。

部署顺序：

1. Stock：发布 `stock:8890`。
2. Item：发布 `item:8891`，连接 Stock。
3. API：加入命名空间，连接 Item。

新增端点后重新部署客户端服务。

## 验证与排错

正常状态：

```text
status = ACTIVE
desiredCount = runningCount
pendingCount = 0
rolloutState = COMPLETED
```

验证顺序：

1. 最新 Task Definition revision 包含命名端口。
2. 应用容器与 Service Connect 代理均为 `RUNNING`。
3. Cloud Map 中存在 `stock` 和 `item`。
4. CloudWatch Logs 无镜像、权限或连接错误。
5. 通过 API 请求验证完整调用链。

出现“客户端和服务器模式至少需要一个端口映射”时，检查：

- ECS Service 是否选择了最新 Task Definition revision。
- Service Connect 页面是否添加了端口，而不只是选择模式。
- `portName` 是否严格匹配 `stock-rpc` 或 `item-rpc`。

`stock:8890` 和 `item:8891` 仅在同一 Service Connect 命名空间的 ECS 任务中有效，不能从本地电脑直接解析；Kitex RPC 也不能用 HTTP `curl` 验证。

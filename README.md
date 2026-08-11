# Kitex Shop

English | [简体中文](./README_CN.md)

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Kitex](https://img.shields.io/badge/Kitex-0.16.3-4C8BF5)](https://www.cloudwego.io/docs/kitex/)
[![Hertz](https://img.shields.io/badge/Hertz-0.10.4-00A6A6)](https://www.cloudwego.io/docs/hertz/)

Kitex Shop is a compact Go microservice sample that demonstrates an HTTP-to-RPC request chain with [CloudWeGo Hertz](https://www.cloudwego.io/docs/hertz/) and [CloudWeGo Kitex](https://www.cloudwego.io/docs/kitex/). It includes Proto3 IDLs, generated Kitex-Protobuf code, multi-stage Docker images, Amazon ECR publishing scripts, and AWS ECS Service Connect deployment definitions.

The project has moved away from etcd-based service discovery. Services now connect through the stable DNS names `item` and `stock`, which are provided by a Docker network during local container runs or by ECS Service Connect in AWS.

## Architecture

```text
Client
  |
  | HTTP GET /api/item
  v
API (Hertz, :8889)
  |
  | Kitex-Protobuf RPC: item:8891
  v
Item Service
  |
  | Kitex-Protobuf RPC: stock:8890
  v
Stock Service
```

| Component | Technology | Port | Responsibility |
| --- | --- | ---: | --- |
| API | Hertz + Kitex client | `8889` | Exposes the HTTP endpoint and calls Item Service |
| Item Service | Kitex server/client | `8891` | Builds item details and requests stock data |
| Stock Service | Kitex server | `8890` | Returns stock data for an item |

The sample endpoint currently requests item ID `1024`. The Item Service returns demonstration metadata, while the Stock Service echoes the item ID as the stock value. This behavior is intentionally simple and is not backed by a database.

## Repository Layout

```text
.
├── api/                         # Hertz HTTP gateway
├── rpc/
│   ├── item/                    # Item Kitex service and Stock client
│   └── stock/                   # Stock Kitex service
├── idl/                         # Proto3 service and message definitions
├── kitex_gen/                   # Code generated from the Proto3 IDLs
├── generate.sh                 # Deterministic Kitex code generation
├── Dockerfile.api               # API multi-stage image
├── Dockerfile.item              # Item Service multi-stage image
├── Dockerfile.stock             # Stock Service multi-stage image
├── aws-ecr-*.sh                 # ECR repository/login/tag/push helpers
├── aws-task-api.json            # API ECS task definition
├── aws-task-item.json           # Item ECS task definition
├── aws-task-stock.json          # Stock ECS task definition
├── aws-task.json                # Standalone Service Connect Nginx example
├── AWS_ECS_SERVICE_CONNECT.md   # ECS Service Connect deployment notes
├── go.mod
└── go.sum
```

Generated files under `kitex_gen/` are committed so the sample builds without requiring Kitex code generation tools.

## Prerequisites

Before building or running the sample, install:

- Go 1.22 or later
- Git
- Docker, for the recommended local end-to-end run

For AWS deployment, you also need:

- AWS CLI v2
- An authenticated AWS account with access to ECR, ECS, IAM, CloudWatch Logs, and Service Connect/Cloud Map
- Docker Buildx or another way to build Linux `amd64` images when deploying from Apple Silicon

## Build from Source

Clone the repository and download dependencies:

```bash
git clone https://github.com/linzworld/kitex-shop.git
cd kitex-shop
go mod download
```

Build all Go packages:

```bash
go build ./...
```

The application source uses `item:8891` and `stock:8890` as service addresses. Those names are resolved automatically in the Docker and ECS workflows below. If you run the Go processes directly on the host, you must provide equivalent DNS or hosts-file mappings yourself.

## Build and Run with Docker

Docker is the simplest local end-to-end option because container names provide the same `item` and `stock` DNS names used by ECS Service Connect.

Build the three images:

```bash
docker build -f Dockerfile.stock -t kitex-shop-stock:latest .
docker build -f Dockerfile.item -t kitex-shop-item:latest .
docker build -f Dockerfile.api -t kitex-shop-api:latest .
```

Create a network and start the services in dependency order:

```bash
docker network create kitex-shop

docker run -d --name stock --network kitex-shop \
  -p 8890:8890 kitex-shop-stock:latest

docker run -d --name item --network kitex-shop \
  -p 8891:8891 kitex-shop-item:latest

docker run -d --name api --network kitex-shop \
  -p 8889:8889 kitex-shop-api:latest
```

Verify the full request chain:

```bash
curl http://localhost:8889/api/item
```

The response is the Kitex string representation of an item containing ID `1024`, title `Kitex`, a sample description, and stock value `1024`.

Inspect logs if a service does not start:

```bash
docker logs stock
docker logs item
docker logs api
```

Remove the local containers and network when finished:

```bash
docker rm -f api item stock
docker network rm kitex-shop
```

## Test

Run the repository-wide Go test command:

```bash
go test ./...
```

The repository includes handler tests for successful responses, required ID validation, Kitex business errors, and the stock fallback behavior. Use the Docker workflow and `curl` command above as an end-to-end smoke test.

## Protobuf APIs

The service contracts live under `idl/`:

- `base.proto` defines the shared `BaseResp` message.
- `item.proto` defines `Item`, `GetItemReq`, `GetItemResp`, and `ItemService.GetItem`.
- `stock.proto` defines the stock request/response messages and `StockService.GetItemStock`.

The services use Kitex-Protobuf over TTHeader/TCP, not standard gRPC. After changing an IDL, install Kitex v0.16.3 and regenerate all committed code with:

```bash
go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3
./generate.sh
go test ./...
```

Do not edit generated files manually unless you are deliberately maintaining a generated-code patch.

## Deploy to AWS ECS

The production-oriented path uses one image and one ECS task definition per application service.

### 1. Build Linux images

The ECR helper scripts expect the local tags shown below. On Apple Silicon, explicitly build for `linux/amd64`:

```bash
docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.stock -t kitex-shop-stock:latest .

docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.item -t kitex-shop-item:latest .

docker buildx build --platform linux/amd64 --load \
  -f Dockerfile.api -t kitex-shop-api:latest .
```

### 2. Publish to ECR

Review `AWS_REGION`, `AWS_ACCOUNT_ID`, and repository names in the scripts before running them:

```bash
./aws-ecr-kitex-shop-stock.sh
./aws-ecr-kitex-shop-item.sh
./aws-ecr-kitex-shop-api.sh
```

Each script verifies the AWS identity, creates the ECR repository if necessary, logs Docker in, tags the local image, pushes it, and prints both tagged and digest image URIs.

### 3. Register task definitions

Review account IDs, role ARNs, regions, image URIs, CPU, memory, and log groups in the JSON files before registration:

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

### 4. Configure Service Connect

Deploy services in this order:

1. Stock publishes `stock:8890` using port name `stock-rpc`.
2. Item publishes `item:8891` using port name `item-rpc` and connects to Stock.
3. API joins the namespace as a client and connects to Item.

Kitex uses its native Protobuf protocol over TTHeader/TCP, so the Item and Stock port mappings must not declare `appProtocol: grpc`. The API port may declare `appProtocol: http`.

See [AWS_ECS_SERVICE_CONNECT.md](./AWS_ECS_SERVICE_CONNECT.md) for namespace, port, logging, security-group, deployment-order, verification, and troubleshooting details.

## Configuration Notes

- Service addresses are currently compiled into the applications as `item:8891` and `stock:8890`.
- The API listens on `0.0.0.0:8889`; Item listens on `0.0.0.0:8891`; Stock listens on `0.0.0.0:8890`.
- The included AWS files contain environment-specific account IDs, regions, ARNs, and log-group names. Treat them as examples and review them before use in another AWS account.
- `aws-task.json` is an independent Nginx Service Connect example and is not part of the Kitex Shop request chain.
- This sample no longer requires etcd and contains no etcd registry client.

## Current Limitations

- The HTTP endpoint uses a fixed item ID and does not accept request parameters.
- Responses use Kitex's string representation rather than a structured JSON API contract.
- There is no database, cache, authentication, retry policy, or application-level observability.
- There are no dedicated unit or integration test suites yet.
- Service endpoints are hard-coded instead of being provided through environment variables or configuration files.

These constraints keep the repository focused on the Hertz-to-Kitex call chain and ECS Service Connect deployment model.

## Related Documentation

- [Kitex documentation](https://www.cloudwego.io/docs/kitex/)
- [Kitex getting started tutorial](https://www.cloudwego.io/docs/kitex/getting-started/tutorial/)
- [Hertz documentation](https://www.cloudwego.io/docs/hertz/)
- [AWS ECS Service Connect documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-connect.html)

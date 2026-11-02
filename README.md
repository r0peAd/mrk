# conncheck

一个用于测试 MySQL、PostgreSQL、Redis、MongoDB 和 Kafka 连接可用性的 Go 命令行工具。配置文件为 YAML，所有数据源均支持多实例。

## 快速开始

```bash
go run ./cmd/conncheck -config configs/config.example.yaml
go run ./cmd/conncheck -config configs/config.example.yaml -output json
go run ./cmd/conncheck -config configs/config.example.yaml -concurrency 2
```

程序并发执行所有检查，输出表格或 JSON。只要有一个实例失败，退出码为 `1`；全部成功时为 `0`；配置或参数错误为 `2`。

## 配置

配置文件包含 `defaults` 和五个数据源数组：

- `mysql`：`host`、`port`、`username`、`password`、`database`、`timeout`、`tls`
- `postgresql`：与 MySQL 相同的基础字段
- `redis`：基础字段；`database` 写成字符串，例如 `"1"`；可用 `sasl` 表示 ACL 用户
- `mongodb`：可用 `host/port` 单节点，也可用 `addresses` 表示副本集或多节点；`attributes.auth_database` 控制认证库
- `kafka`：必须配置 `brokers`，可启用 `tls` 和 `sasl`（`plain`、`scram-sha-256`、`scram-sha-512`）

`defaults` 中的 `username`、`password`、`database` 和 `timeout` 会应用到所有未显式配置的实例。完整示例见 `configs/config.example.yaml`。

## 开发

```bash
go test ./...
go build ./cmd/conncheck
make build
make build-all
make checksums
```

`make build-all` 会在 `build/` 下生成交叉编译产物，支持 `darwin/amd64`、`darwin/arm64`、`linux/amd64`、`linux/arm64`、`linux/386`、`linux/arm`、`windows/amd64`、`windows/arm64`。Windows 产物为 `.exe`，其他系统为同名可执行文件。交叉编译统一使用 `CGO_ENABLED=0`。

## 注意

- Kafka 目前使用第一个 broker 建立连接并读取元数据；其余 broker 仅保留在配置中。
- 生产配置中包含敏感信息时，请避免提交到 Git，或改为通过模板生成。

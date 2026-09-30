# Changelog

- Add the backward-compatible `ConnectorSessionPolicyScheduledReuse` v1 constant
  for demand-aware industrial sessions; no DTO, RPC, or interface shape changed.

本文档记录 `iot-rule-plugin-sdk` 的公共 API 与工程结构演进。

## v1.0.0 - 2026-09-03

### Connector lifecycle contract (2026-09-08)

- 新增 `per_operation`、`persistent`、`pooled`、`stateless`、`idle_timeout`
  Connector 会话策略常量。
- 不修改 `ConnectorResource` 字段、接口方法集或 RPC DTO，保持已发布 v1 插件
  二进制和线协议兼容。
- Connector descriptor 使用 `session:<policy>` capability 声明支持的生命周期；
  插件进程生命周期始终与外部设备连接生命周期分离。
- 新增 `docs/connector-session-lifecycle.md` 和契约常量测试。

### Reliability hardening (2026-09-05)

- 完整冻结根 API、schema、接口方法集、结构体字段/标签、常量与 RPC 线协议。
- RPC 超时会关闭失效连接，所有关键 Engine、Plugin 与 Connector 调用均有 deadline。
- 健康状态使用最新状态必达槽位、发送重试和可控排空；新增 Flush 与统计辅助函数。
- 无 Connector 的默认客户端改为返回 `ErrNotImplemented`。
- SDK StreamConsumer 支持并发调用和幂等关闭。
- Event ID 保持 `evt_<纯数字>` 形式并提供进程内单调唯一性。
- 内部插件启动与 Engine 加载使用固定握手副本，不受兼容导出变量修改影响。
- 新增 CI、LICENSE、SECURITY、CONTRIBUTING 和运行时契约文档。

### Changed

- 项目正式命名为 IoT Rule Engine Plugin SDK。
- Go module 路径变更为 `github.com/punk-one/iot-rule-plugin-sdk`。
- 公共 API 以原 `punk-rule-plugin-sdk v1.8.1` 为兼容基线。
- 保留既有插件握手值，以兼容迁移前构建的插件二进制。

### Validation

- `GOWORK=off go test ./...`

以下内容为原 SDK 的继承历史。

## v1.8.1 - 2026-05-14

### Changed

- `ConnectorBindingSpec.Required` 现在总是显式出现在 JSON 中，避免 `required:false` 在契约序列化时被省略。
- README 与 processor 示例已明确 `processor` 可以声明单 connector 绑定契约。

### Added

- 新增 `processor` connector 契约回归测试，锁定 `PluginTypeProcessor` 的 `connector_binding` JSON 形态。

### Validation

- `go test ./...`

## v1.8.0 - 2026-04-20

### Changed

- `StartArgs` 新增 `ResourceEventsEnabled`，由宿主显式声明当前节点是否具备资源事件能力。
- `RuntimeContext.ResourceEvents()` 改为懒启动语义：首次调用时才启动资源事件轮询。
- 当节点不具备资源事件能力时，SDK 不再后台轮询 `NextResourceEvent`。
- 资源事件轮询新增异常快返回退避与指标观测，避免 RPC/Engine 回归时形成 CPU busy loop。

### Validation

- `go test ./...`

## v1.7.3 - 2026-04-19

### Added

- `StreamOpenRequest` 新增通用流控字段：
  - `MaxBufferedMessages`
  - `OverflowPolicy`
  - `AckTimeoutMS`
  - `MaxDeliverBatch`

### Changed

- RuntimeContext 的 connector stream consumer 现在支持按 `MaxDeliverBatch` 批量接收并在本地缓冲，减少高频 Source 的 RPC 往返开销。
- 流式 connector 的公共契约现在明确区分：
  - ingress/backlog
  - delivery credits
  - overflow policy

### Validation

- `go test ./internal/runtime`

## v1.7.2 - 2026-04-18

### Added

- 新增通用流式 connector 会话公共模型：
  - `StreamConsumer`
  - `StreamOpenRequest`
  - `StreamOpenResponse`
  - `StreamReceiveRequest`
  - `StreamReceiveResponse`
  - `StreamMessage`
  - `StreamAckRequest`
  - `StreamNackRequest`
  - `StreamGrantCreditsRequest`
  - `StreamCloseRequest`
- 新增流式投递语义常量：
  - `StreamDeliveryBestEffort`
  - `StreamDeliveryAtLeastOnce`

### Changed

- `ConnectorClient` 现在支持 `OpenStream(...)`，用于 Source 插件消费原生事件流。
- `ConnectorPlugin` 现在支持完整流式生命周期：
  - `OpenStream`
  - `ReceiveStream`
  - `AckStream`
  - `NackStream`
  - `GrantCredits`
  - `CloseStream`
- Engine RPC / Connector RPC / RuntimeContext 已扩展流式桥接能力。
- `PUBLIC_API` 基线已扩展到 connector stream contract。

### Validation

- `GOWORK=off go test ./...`

## v1.7.1 - 2026-04-04

### Added

- 新增 connector 契约公开类型：
  - `ConnectorDescriptor`
  - `ConnectorBindingSpec`

### Changed

- `PluginCapabilities` 新增两个可选字段：
  - `ConnectorDescriptor *ConnectorDescriptor`
  - `ConnectorBinding *ConnectorBindingSpec`
- README 已补充 connector family / capability / binding 的正式声明示例。
- `PUBLIC_API` 基线已扩展到 connector compatibility contract。

### Validation

- `go generate ./...`
- `go test ./...`

## v1.7.0 - 2026-04-04

### Added

- 新增 connector 插件稳定入口：
  - `PluginTypeConnector`
  - `ConnectorPlugin`
  - `ConnectorClient`
  - `ServeConnector(...)`
- 新增资源运行时能力：
  - `RuntimeContext.Connector()`
  - `RuntimeContext.ResourceEvents()`
- 新增 connector 公开模型：
  - `ConnectorResource`
  - `ConnectorRequest`
  - `ConnectorResponse`
  - `ResourceStatusEvent`
  - `ProviderPolicy`
  - `WarmupPolicy`
- 新增正式策略类型：
  - `HealthPolicy`
  - `QuotaPolicy`
- 新增 `ConnectorResource` 策略 helper：
  - `SetHealthPolicy(...)`
  - `GetHealthPolicy()`
  - `SetQuotaPolicy(...)`
  - `GetQuotaPolicy()`
- 新增 connector 公开 API 黑盒测试：`tests/connector_resource_test.go`

### Changed

- `PUBLIC_API` 基线已扩展到 connector-aware 2.x 运行时。
- `PluginRuntimeContext` 现在会为外部插件桥接 connector 调用与资源状态事件流。
- README、公开 API 基线与测试已对齐 connector 插件开发模式。

### Validation

- `go test ./tests ./internal/runtime` 已通过。

## v1.6.0 - 2026-03-27

### Added

- 新增 `BasePlugin`，为插件作者提供无副作用生命周期默认实现。
- 新增结构化错误模型：`PluginError`、`ErrorKind`、`ErrNotImplemented`、`ErrBatchNotSupported`。
- 新增事件工具：`Event.Clone()`、`GetTraceID(...)`、`SetTraceID(...)`。
- 新增 API 治理文件：`public_api.json`、`PUBLIC_API.md`、`internal/tools/public_api_export.go`。
- 新增健康接入文档：`docs/health_reporter.md`。
- 新增可运行示例：`examples/health_basic/main.go`。
- 新增维护入口：`doc.go`，支持 `go generate ./...` 刷新公开 API 文档。

### Changed

- `PluginRPC` 现在支持结构化错误透传，不再丢失重试/致命/批量不支持语义。
- 批量分发链路支持自动回退：未声明 `SupportBatch` 或运行时返回 `ErrBatchNotSupported` 时，自动退回逐条 `OnEvent(...)`。
- `ServeOptions.Health` 现在会参与插件启动时的默认健康配置合并。
- `Stop()` 失败时的健康上报消息改为原始错误文本，不再暴露内部 RPC 编码串。
- README、示例、测试和公开 API 基线已对齐。

### Refactored

- 根目录职责收敛为“稳定公共 API”；示例统一放入 `examples/`，黑盒测试统一放入 `tests/`。
- `runtime_context`、`emitter`、`metrics`、`engine_rpc` 已按“接口 / 实现 / RPC 适配”拆分为更清晰的文件。

### Removed

- 从根包移除实现细节导出：`PluginRuntimeContext`、`NewPluginRuntimeContext`、`EngineRPCEmitter`、`EngineRPCLogger`、`EngineRPCMetrics`、`NoOpEmitter`、`StateManagerImpl`。
- SDK 自身测试改为优先依赖稳定接口，运行时实现测试下沉到 `internal/runtime`。

### Validation

- `go test ./...` 已通过。

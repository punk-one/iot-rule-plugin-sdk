# iot-rule-plugin-sdk v1 三年兼容政策

## 1. 承诺期

`iot-rule-plugin-sdk` 自 `v1.0.0` 首次公开发布日起至少 36 个月保持 `v1.x.y`，不发布要求现有插件迁移的 `v2`。若 `v1.0.0` 于 2026-09-04 公开发布，则最低承诺期截止到 2029-09-04。

版本规则：

- `v1.x.y` 的 patch：缺陷、安全、性能和文档修复，不增加使用方必须实现的能力。
- `v1.x.0` 的 minor：只增加可选能力，旧插件不修改即可继续编译和运行。
- 承诺期内不删除已公开符号、RPC 方法、字段、常量、握手值或既有行为语义。
- 废弃 API 可以标记 deprecated，但承诺期内保留实现；移除只能进入承诺期后的大版本。

## 2. 兼容范围

### 2.1 Go 源码兼容

以下内容一经进入稳定版即冻结：

- 导出函数和方法的名称、参数、返回值与泛型约束；
- `Plugin`、`ConnectorPlugin`、`RuntimeContext`、`Logger`、`Emitter`、`Metrics`、`StateManager`、`StateStore`、`ConnectorClient`、`StreamConsumer`、`HealthReporter` 的方法集；
- 导出结构体的字段名、字段类型和标签；
- 导出常量的类型与值；
- 错误分类及 `errors.Is/As` 行为；
- 根 module path `github.com/punk-one/iot-rule-plugin-sdk`。

不得给由插件作者实现的既有接口添加方法。新能力必须定义新的最小可选接口，并通过 capability 或类型断言协商。

不得假设给导出结构体增加字段永远兼容：使用无字段名复合字面量的代码会因此编译失败。新增数据优先使用新类型、Options、版本化 payload，或先明确禁止外部使用无字段名字面量。

### 2.2 RPC 线协议兼容

以下值在整个 v1 承诺期内固定：

- HashiCorp go-plugin `ProtocolVersion = 1`
- `MagicCookieKey = PUNK_PLUGIN_MAGIC`
- `MagicCookieValue = punk-rule-engine`
- dispense key：`plugin`、`connector`
- net/rpc 服务与方法名
- 已有 DTO 的 Gob 可见字段名、类型和含义
- 结构化错误前缀及已公开错误 kind

可以新增 RPC 方法，但旧引擎和旧插件不得被迫调用。新方法必须由 capability 协商，并为“不支持”提供结构化回退。

### 2.3 行为兼容

下列内容属于契约，而非内部实现细节：

- Batch 支持声明与逐条回退规则；
- retryable、skippable、fatal、not_implemented、batch_unsupported 的含义；
- Event 的 JSON 字段名、时间格式、metadata trace key；
- Ack/Nack、credits 与 at-least-once 的行为；
- nil、空切片、空 map、零值 Options 的默认语义；
- 健康状态、状态去重和关闭期间的交付保证；
- Connector family/capability 匹配规则。

缺陷修复若会改变可观察行为，必须在 changelog 中单列“行为变更”，并增加回归 fixture。

### 2.4 二进制兼容

插件是独立可执行文件，经 go-plugin/net-rpc 通信，不承诺 Go 进程内 ABI 兼容；承诺的是旧插件可执行文件的进程启动、握手和 RPC 兼容。

最低矩阵：

| 方向 | 必须验证 |
| --- | --- |
| 旧插件 → 新引擎 | 用 `v1.0.0` SDK 构建的冻结二进制可被每个后续 v1 引擎加载并执行基础用例 |
| 新插件 → 旧引擎 | 只使用 v1.0 基础能力的新插件可被最老受支持 v1 引擎加载；新能力在旧引擎上明确降级 |
| 当前源码 | 当前 SDK 示例和官方插件全部编译、普通测试与竞态测试通过 |

冻结 fixture 应包含 Source、Processor、Sink、Connector、批处理、结构化错误、State、Health、Stream 各一条最小路径，并保存构建环境与 SHA-256。

## 3. 扩展规则

允许：

- 新增独立函数、常量、新类型和新的可选小接口；
- 新增 capability 值，并让未知值可被旧宿主忽略；
- 为现有字符串策略字段新增标准值；Connector 会话生命周期值及兼容规则见
  `connector-session-lifecycle.md`；
- 新增 RPC 方法，前提是旧端不必调用且有协商；
- 优化内部实现，只要公开行为和线格式不变；
- 增加更严格的输入校验，但必须评估现有合法输入是否被拒绝。

禁止：

- 向已有公开接口追加方法；
- 重命名或删除导出符号；
- 修改 DTO 已有字段类型、JSON/Gob 名称或复用旧字段表达新含义；
- 修改握手值、RPC 服务名和既有方法名；
- 改变默认值而不提供兼容开关；
- 将原本同步成功的操作改成静默丢弃，或改变 Ack/重试语义；
- 修改已发布 Git 标签。

## 4. 发布门禁

每次 v1 release 必须同时通过：

1. `go test ./...`
2. `go test -race ./...`
3. `go vet ./...`
4. API 完整签名 diff，无未批准的 breaking change
5. RPC/golden fixture diff，无未批准的 wire change
6. 旧插件二进制兼容矩阵
7. 官方插件构建与契约测试
8. 依赖漏洞扫描与许可证检查
9. changelog、迁移说明和 deprecated 列表检查
10. release artifact checksum、SBOM 与签名

任一门禁失败不得创建正式 tag。

## 5. 支持和废弃

- 维护最新 minor，并对仍在支持矩阵内的旧 minor 提供兼容验证。
- 安全和数据丢失类缺陷优先发布 patch。
- 废弃至少提前 12 个月公告，并提供替代 API、示例和自动迁移建议。
- 三年期满不代表立即发布 v2；只有存在无法通过兼容扩展解决的需求时才启动 v2。
- 若严重安全问题确实无法兼容修复，安全优先，但必须发布安全公告、影响矩阵、迁移工具和并行兼容窗口，不得静默破坏。

## 6. 变更审批清单

每个涉及 SDK 契约的 PR 必须回答：

- 是否改变公开函数/方法签名或接口方法集？
- 是否改变结构体字段、标签、常量、错误 kind？
- 是否改变握手、RPC 方法或 Gob/JSON 线格式？
- v1.0.0 旧插件二进制是否仍通过？
- 新能力在旧引擎上如何检测和降级？
- 是否改变超时、重试、Ack、队列、关闭或默认值语义？
- 是否更新兼容 fixture、文档和 changelog？

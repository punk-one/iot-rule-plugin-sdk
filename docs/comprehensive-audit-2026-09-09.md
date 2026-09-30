# IoT Rule Plugin SDK 当前实现综合审计（2026-09-09）

## 基线与结论

基线为本地 HEAD `abab669466a4d8b2bb5974886e8ae200f09e7653` 加当前全部未提交变更；HEAD 不包含全部审计实现。环境为 Go 1.25.0、WSL Linux amd64。重点复核公开契约、Event 所有权、日志与 Context 生命周期、RPC 超时和三年兼容承诺。

**结论：源码与竞态门禁通过，但不宜直接冻结为三年稳定基线。** 日志协程泄漏、Clone 的部分行为差异均已复现，应在发布稳定基线前处理。修复可以优先采用内部实现和可选接口，避免扩大既有 Plugin/RuntimeContext/Logger 等接口的方法集。

此次为代码审查与隔离诊断，不是依赖漏洞全扫描、跨平台认证或三年可用性保证；没有修改公共 API、生成契约基线或生产实现。

## 验证记录

| 检查 | 本次结果 |
| --- | --- |
| `go test -race -count=1 ./...` | 通过，禁用测试缓存，包含 API/线协议契约测试 |
| `go vet ./...` | 通过 |
| `go build ./examples/...` | 通过 |
| Clone、Context 生命周期诊断 | 三个行为问题成功复现 |
| 旧版冻结二进制互操作 | 本轮未运行；当前 CI 也未看到冻结旧插件矩阵 |
| 生成文档重写、标签发布 | 未执行；避免覆盖未提交的公开 API 基线 |

复现文件：[ownership_logger_repro.go](audit-repros-2026-09-09/ownership_logger_repro.go)，带 `//go:build ignore`，不参与普通构建。从 SDK 根目录运行：

```bash
go run docs/audit-repros-2026-09-09/ownership_logger_repro.go
```

本轮输出：

```text
empty_slice_before={"empty":[]} after={"empty":null}
pointer_map_original_after_clone_mutation=2
logger_flushers_retained_after_20_context_closes=20
```

## 发现与建议

### SDK-A01 · P1 · Context 关闭不会停止日志刷新协程（已复现）

**证据：** `internal/runtime/logger.go:37` 在创建 Logger 时启动 goroutine；`:41` 的 startFlusher 永久遍历 ticker，没有关闭信号。每个 Context 在 `context.go:285` 创建一个 Logger，而 `:336` 的 Close 只关闭 resource event 的 stopCh。

创建并关闭 20 个 Context 后，通过 goroutine 堆栈计数仍新增 20 个 startFlusher。生产插件进程长期复用、规则频繁部署/取消部署时会累积 goroutine，并保留 logger/bridge 引用。复现只证明日志协程保留，没有把全部 TCP 连接泄漏作为已验证事实。

**建议：** 内部 logger 增加 stop/once/done、有限时间 flush；Context.Close 统一关闭自己拥有的 logger 与运行时资源，所有失败路径也关闭。可使用内部可选 Close 接口，保持公共 Logger 方法集不变。
**验收：** 1000 次 Context 创建/关闭后日志协程回到基线；并发关闭、Start 失败、上报阻塞、重复 Close 均不泄漏或死锁。

**相关问题：** `logger.go:49` 在持锁期间同步 LogBatch；第 100 条日志也在调用线程触发 Flush，不是完全异步的有界日志队列。fields 仅浅拷贝，其嵌套可变值可在异步 flush 前被调用方修改。应在队列边界明确字段所有权、限制单条/总字节，并把 RPC 放到锁外；不能仅增加 mutex。

### SDK-A02 · P1 · “JSON 兼容”类型范围未闭合，指针容器仍共享（已复现，受输入类型约束）

**证据：** `internal/core/event.go:148` 默认回退调用反射复制，`:165` 只处理 Slice/Array/Map，其他类型原样返回。指向 map 的指针可以被 encoding/json 编码，但 Clone 后仍指向同一个 map；含 slice/map 的普通 struct 也不会递归复制。

规范 `docs/runtime-contracts.md` 只表述“JSON 兼容值”，没有明确“仅标准 JSON 树类型、禁止指针/结构体”，也没有与 Clone 绑定的边界验证。本轮 pointer-map 复现中，修改克隆值后原始 map 的 n 从 1 变成 2。

常规 RPC 已通过 JSON 编码/解码时，指针会被归一化，不能据此断言所有跨进程事件都共享内存；风险主要位于 SDK 使用方、本地异步分支和引擎进程内调用。Clone 仍需要调用方保证复制期间不并发写原值。

**建议：** 发布前明确允许的类型集合：标准 JSON map/slice/scalar 与明确列出的 byte 表示；对指针/结构体选择安全支持或在进入 Clone 前拒绝。增加深度、容量和循环引用防护。若扩大复制范围，避免更改合法数值类型。
**验收：** 所有允许类型都不共享可变对象；不允许的值明确报错；嵌套命名容器、指针/结构体、循环引用和过深结构有覆盖。循环引用属于非 JSON 输入，本轮未执行会导致堆栈耗尽的攻击。

### SDK-A03 · P2 · Clone 把非 nil 空切片改为 nil（已复现）

**证据：** `internal/core/event.go:156` 的 cloneSlice 使用 `append([]T(nil), values...)`。对非 nil 但长度为 0 的 []string 等输入，返回 nil；[]byte 的类似分支也会改变 nil/empty 区别。

JSON 会从 `[]` 变为 `null`，可能影响下游 schema、条件判断和签名。规范 `docs/v1-compatibility-policy.md` 明确把 nil/empty 语义列入 v1 契约，因此属于行为兼容缺陷。

**建议：** 非 nil 输入用 make+copy 保留非 nil 空集合；检查 []byte、RawMessage、map 等全部分支的 nil/empty 策略，统一文档。
**验收：** 对允许类型 Clone 前后 JSON 语义和 Go nil 性保持一致，数值类型不变，并保证深拷贝隔离。

### SDK-A04 · P2 · 观测 RPC 超时会关闭业务共用 bridge（静态确认的架构风险）

**证据：** `internal/transport/engine_rpc.go:25,46,65` 中 Emit、LogBatch、Health 共用同一 rpc.Client；`rpc_call.go` 超时后关闭整个 client。日志/指标错误在 void 方法中被丢弃。

一次日志/健康上报阻塞可能导致同实例后续 Emit、State、Connector 请求也失败。该行为已经写进 runtime-contracts 的“宿主必须替换 client”说明，不是未公开的行为变更；但其恢复路径需要端到端验收，不能只测 timeout 返回错误。

**建议：** 完整验证宿主失效检测/实例重建；若隔离观测通道，作为可选能力协商，保留旧版本回退。统计丢失日志与 bridge 损坏，防止静默。
**验收：** 用故障注入阻塞 LogBatch，观察业务调用受影响范围、恢复时限、事件失败是否正确反馈；正常 Stop 与其他规则实例不受误杀。

### SDK-A05 · P2 · 三年兼容政策已有，冻结二进制执行矩阵尚未落地（静态确认）

**证据：** `docs/v1-compatibility-policy.md` 明确要求旧插件→新引擎、新插件→旧引擎及冻结 Source/Processor/Sink/Connector fixture；当前 `.github/workflows/ci.yml` 执行当前源码测试、race、vet、示例构建和生成文档 diff，没有运行这些旧二进制。

当前 contract_v1 对函数/类型/常量/线协议提供比单纯符号清单更完整的快照；因此不能把 API 检测概括成“仅名称检查”。但静态签名无法证明旧二进制生命周期、默认语义和失败恢复兼容。

**建议：** 正式 v1.0.0 发布时冻结二进制、构建环境及 SHA-256，以不可变版本运行双向矩阵；支持日期以首次公开发布为准，不把文档示例日期当作已兑现的公开发布日期。
**验收：** 每次 release CI 必跑旧版 fixture、可选能力回退、SDK-A03 行为 golden tests；补齐受支持 Go/OS 范围，发布失败禁止打正式 tag。

## 已确认的基础能力

- 普通 map/slice 和常用 typed slice 已有深拷贝实现，事件 ID 用原子递增规避同进程碰撞。
- ServeFactory 每次 Dispense 创建独立对象，适合同 artifact 进程多规则；源码兼容保留的 Serve 不等于工厂隔离。
- RPC 默认有限 timeout，超时关闭连接的行为已有明确契约；不再把 net/rpc 超时后的每次 goroutine 泄漏作为现有结论。
- StreamConsumer 有 Recv 串行、Close 幂等和 closed 状态；相关 race 测试通过。
- v1 政策已明确最低 36 个月与不能向既有接口追加必选方法。当前 AGPL-3.0-only 文档一致，许可证选择不再待定。
- SDK 的缺陷会传导到 Engine 和全部共享该 SDK 的插件，修复后需要一并重建与验证，不能只替换 SDK 源文件。

## 整改与兼容验收顺序

1. 修 SDK-A01，验证实例启停后资源回收，优先消除三年维护期内累积性故障。
2. 在冻结 v1.0.0 前明确 SDK-A02 输入集合，修 SDK-A03；如果已分发版本，单列行为修复说明与兼容样例。
3. 验证 SDK-A04 的宿主恢复闭环，再落地 SDK-A05 冻结二进制矩阵。
4. SDK 修复通过后，重建官方 23 个插件并运行 Engine 多规则、背压、停止与重启集成回归。
5. 完成依赖安全扫描、来源/许可证检查和不可变 release 基线。

以上是审计发现与建议，生产代码在本轮未被修复；当前测试全绿不能关闭这些问题。

# v1 运行时契约

## 单进程多实例

- 普通插件使用 `ServeFactory`，同一个插件进程可接受多次 `plugin_factory` Dispense。
- 每次 Dispense 必须创建独立 Plugin 对象；不同规则节点不得共享配置、Context 或 Stop 状态。
- `StartArgs.BrokerID` 标识本实例的反向 Engine RPC；零值仅用于兼容旧宿主并回退到 3。
- Connector 使用 `ServeConnector`，在单进程中通过不同 `provider_handle` 隔离资源。
- `Serve` 保留源码兼容，但不具备强制工厂隔离语义，不应接入共享进程池。

## RPC deadline

- Info、Start、Stop、事件、状态和 Connector 控制调用默认最长 30 秒。
- 日志、指标和健康上报最长 5 秒。
- Connector 请求显式配置 `TimeoutMS` 时，RPC deadline 为业务超时加 1 秒传输余量。
- Stream Receive 的 deadline 为 `WaitTimeoutMS` 加 1 秒传输余量。
- `net/rpc` 不支持取消单个调用。超时时 SDK 会关闭整个失效连接以释放挂起调用；
  宿主必须销毁并重建对应插件或 Engine bridge，不能继续复用该 client。

## StreamConsumer 并发

SDK 提供的 StreamConsumer 支持并发调用。多个 `Recv` 会被串行化以保护本地
批量缓冲；`Close` 幂等并可与 `Recv`、Ack、Nack、GrantCredits 并发。Close
开始后，新调用会明确失败。Connector 服务端实现仍需自行保护其 provider handle。

## Event 所有权

- `OnEvent`/`OnEvents` 收到的 Event 由本次插件调用独占。
- 插件把同一事件交给多个异步任务前，必须为每个任务调用 `Clone()`。
- Emitter/Engine 在进程与调度边界再次取得所有权。
- Payload 只允许 JSON 兼容值；`[]byte` 和 `json.RawMessage` 会复制底层数组。
- Event ID 是以 `evt_` 开头的不透明标识。当前实现保持纯数字后缀，但使用方不得
  从中推导时间或顺序。

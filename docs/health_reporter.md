# HealthReporter 交付与关闭语义

`RuntimeContext.Health()` 返回稳定的 `HealthReporter` 接口。v1 不会向该接口
追加方法；排空和统计通过独立辅助函数提供。

## 状态与心跳

- 状态变化进入有界队列。队列满后保留“最新状态”槽位，新的状态替换该槽位，
  因此最终状态不会因为瞬时突发静默丢失。
- 心跳允许合并，只保留尚未发送的最新心跳。
- 发送失败会按有界退避重试；只有成功发送后才更新“已发送状态”和去重时间。
- 相同状态在 `MaxSilencePeriod` 内会被抑制；状态上报也会重置心跳间隔。
- `Report`、`Healthy`、`Degraded`、`Recovering`、`Unhealthy` 在 Reporter
  关闭后返回错误。

## 排空

插件的 `Stop` 流程会关闭 Reporter，并在最多 10 秒内排空已接受的状态。
需要在业务流程中主动等待时可调用：

```go
err := sdk.FlushHealthReporter(ctx.Health(), 3*time.Second)
```

不得把 Flush 当作网络可用性的保证；超时表示调用方必须按停止或故障策略处理。

## 可观测统计

```go
stats := sdk.GetHealthRuntimeStats(ctx.Health())
```

统计包括状态合并数、心跳合并数和 Sender 错误数。它们是进程内单调计数，适合
诊断队列压力和 Engine RPC 故障。

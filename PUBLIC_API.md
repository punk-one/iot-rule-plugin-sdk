# Public API Surface

Generated from `public_api.json` via `go run ./internal/tools/public_api_export.go`.

## Machine-readable Source

```json
{
  "symbols": [
    "AckArgs",
    "AckMessage",
    "AckStatusFailed",
    "AckStatusSuccess",
    "BasePlugin",
    "BuiltinRuntimeContext",
    "ConnectorBindingSpec",
    "ConnectorClient",
    "ConnectorDescriptor",
    "ConnectorPlugin",
    "ConnectorRPC",
    "ConnectorRPCClient",
    "ConnectorRPCServer",
    "ConnectorRequest",
    "ConnectorRequestBatchRead",
    "ConnectorRequestBatchWrite",
    "ConnectorRequestKind",
    "ConnectorRequestMetadata",
    "ConnectorRequestRead",
    "ConnectorRequestWrite",
    "ConnectorResource",
    "ConnectorResponse",
    "ConnectorSessionPolicyIdleTimeout",
    "ConnectorSessionPolicyPerOperation",
    "ConnectorSessionPolicyPersistent",
    "ConnectorSessionPolicyPooled",
    "ConnectorSessionPolicyScheduledReuse",
    "ConnectorSessionPolicyStateless",
    "DecodeEvent",
    "DefaultHandshakeConfig",
    "DefaultHealthOptions",
    "DefaultLogger",
    "DefaultMetrics",
    "EdgeSubject",
    "EmitBatchArgs",
    "Emitter",
    "EmitterAdapter",
    "EncodeEvent",
    "EngineBridgeRuntimeContext",
    "EngineRPC",
    "EngineRPCClient",
    "ErrBatchNotSupported",
    "ErrNotImplemented",
    "ErrorKind",
    "ErrorKindBatchUnsupported",
    "ErrorKindFatal",
    "ErrorKindNotImplemented",
    "ErrorKindRetryable",
    "ErrorKindSkippable",
    "Event",
    "EventMetadataTraceIDKey",
    "FatalError",
    "Field",
    "FlushHealthReporter",
    "GetHealthRuntimeStats",
    "GetTraceID",
    "HandshakeConfig",
    "HealthDegraded",
    "HealthHealthy",
    "HealthHeartbeat",
    "HealthInitializing",
    "HealthKind",
    "HealthKindHeartbeat",
    "HealthKindState",
    "HealthOptions",
    "HealthPolicy",
    "HealthRecovering",
    "HealthReport",
    "HealthReporter",
    "HealthRuntimeStats",
    "HealthStatus",
    "HealthStopped",
    "HealthUnhealthy",
    "HealthUnknown",
    "InfoReply",
    "InitArgs",
    "InitReply",
    "IsBatchNotSupported",
    "IsFatal",
    "IsRetryable",
    "IsSkippable",
    "LogBatchArgs",
    "LogLevel",
    "LogLevelDebug",
    "LogLevelError",
    "LogLevelInfo",
    "LogLevelWarn",
    "Logger",
    "MetricArgs",
    "Metrics",
    "NewEngineRPCClient",
    "NewEvent",
    "NewFatalError",
    "NewRPCEmitterFromEngine",
    "NewRPCLoggerFromEngine",
    "NewRPCMetricsFromEngine",
    "NewRetryableError",
    "NewSkippableError",
    "NewStateManager",
    "Plugin",
    "PluginCapabilities",
    "PluginConfig",
    "PluginError",
    "PluginFactory",
    "PluginInfo",
    "PluginRPC",
    "PluginRPCClient",
    "PluginRPCServer",
    "PluginType",
    "PluginTypeConnector",
    "PluginTypeProcessor",
    "PluginTypeSink",
    "PluginTypeSource",
    "PluginTypeUtility",
    "ProviderPolicy",
    "PublishAckArgs",
    "QuotaPolicy",
    "ReceiveEventArgs",
    "ReceiveEventReply",
    "ReceiveEventsArgs",
    "ReceiveEventsReply",
    "ReportHealthArgs",
    "ResourceStatus",
    "ResourceStatusBlocked",
    "ResourceStatusCongested",
    "ResourceStatusDegraded",
    "ResourceStatusDisconnected",
    "ResourceStatusEvent",
    "ResourceStatusHealthy",
    "ResourceStatusReadonly",
    "ResourceStatusRecovering",
    "ResourceStatusUnknown",
    "RetryableError",
    "RuntimeContext",
    "RuntimeContextImpl",
    "Serve",
    "ServeConnector",
    "ServeFactory",
    "ServeOptions",
    "SetTraceID",
    "SkippableError",
    "StartArgs",
    "StartReply",
    "StateGetReply",
    "StateKeyArgs",
    "StateManager",
    "StateSetArgs",
    "StateSetWithTTLArgs",
    "StateStore",
    "StatefulRuntimeContext",
    "StopReply",
    "StreamAckRequest",
    "StreamCloseRequest",
    "StreamConsumer",
    "StreamDeliveryAtLeastOnce",
    "StreamDeliveryBestEffort",
    "StreamDeliverySemantics",
    "StreamGrantCreditsRequest",
    "StreamMessage",
    "StreamNackRequest",
    "StreamOpenRequest",
    "StreamOpenResponse",
    "StreamReceiveRequest",
    "StreamReceiveResponse",
    "WarmupPolicy",
    "WarmupPolicyAlwaysOn",
    "WarmupPolicyLazy",
    "WarmupPolicyPreload"
  ]
}
```

## Symbols

- `AckArgs`
- `AckMessage`
- `AckStatusFailed`
- `AckStatusSuccess`
- `BasePlugin`
- `BuiltinRuntimeContext`
- `ConnectorBindingSpec`
- `ConnectorClient`
- `ConnectorDescriptor`
- `ConnectorPlugin`
- `ConnectorRPC`
- `ConnectorRPCClient`
- `ConnectorRPCServer`
- `ConnectorRequest`
- `ConnectorRequestBatchRead`
- `ConnectorRequestBatchWrite`
- `ConnectorRequestKind`
- `ConnectorRequestMetadata`
- `ConnectorRequestRead`
- `ConnectorRequestWrite`
- `ConnectorResource`
- `ConnectorResponse`
- `ConnectorSessionPolicyIdleTimeout`
- `ConnectorSessionPolicyPerOperation`
- `ConnectorSessionPolicyPersistent`
- `ConnectorSessionPolicyPooled`
- `ConnectorSessionPolicyScheduledReuse`
- `ConnectorSessionPolicyStateless`
- `DecodeEvent`
- `DefaultHandshakeConfig`
- `DefaultHealthOptions`
- `DefaultLogger`
- `DefaultMetrics`
- `EdgeSubject`
- `EmitBatchArgs`
- `Emitter`
- `EmitterAdapter`
- `EncodeEvent`
- `EngineBridgeRuntimeContext`
- `EngineRPC`
- `EngineRPCClient`
- `ErrBatchNotSupported`
- `ErrNotImplemented`
- `ErrorKind`
- `ErrorKindBatchUnsupported`
- `ErrorKindFatal`
- `ErrorKindNotImplemented`
- `ErrorKindRetryable`
- `ErrorKindSkippable`
- `Event`
- `EventMetadataTraceIDKey`
- `FatalError`
- `Field`
- `FlushHealthReporter`
- `GetHealthRuntimeStats`
- `GetTraceID`
- `HandshakeConfig`
- `HealthDegraded`
- `HealthHealthy`
- `HealthHeartbeat`
- `HealthInitializing`
- `HealthKind`
- `HealthKindHeartbeat`
- `HealthKindState`
- `HealthOptions`
- `HealthPolicy`
- `HealthRecovering`
- `HealthReport`
- `HealthReporter`
- `HealthRuntimeStats`
- `HealthStatus`
- `HealthStopped`
- `HealthUnhealthy`
- `HealthUnknown`
- `InfoReply`
- `InitArgs`
- `InitReply`
- `IsBatchNotSupported`
- `IsFatal`
- `IsRetryable`
- `IsSkippable`
- `LogBatchArgs`
- `LogLevel`
- `LogLevelDebug`
- `LogLevelError`
- `LogLevelInfo`
- `LogLevelWarn`
- `Logger`
- `MetricArgs`
- `Metrics`
- `NewEngineRPCClient`
- `NewEvent`
- `NewFatalError`
- `NewRPCEmitterFromEngine`
- `NewRPCLoggerFromEngine`
- `NewRPCMetricsFromEngine`
- `NewRetryableError`
- `NewSkippableError`
- `NewStateManager`
- `Plugin`
- `PluginCapabilities`
- `PluginConfig`
- `PluginError`
- `PluginFactory`
- `PluginInfo`
- `PluginRPC`
- `PluginRPCClient`
- `PluginRPCServer`
- `PluginType`
- `PluginTypeConnector`
- `PluginTypeProcessor`
- `PluginTypeSink`
- `PluginTypeSource`
- `PluginTypeUtility`
- `ProviderPolicy`
- `PublishAckArgs`
- `QuotaPolicy`
- `ReceiveEventArgs`
- `ReceiveEventReply`
- `ReceiveEventsArgs`
- `ReceiveEventsReply`
- `ReportHealthArgs`
- `ResourceStatus`
- `ResourceStatusBlocked`
- `ResourceStatusCongested`
- `ResourceStatusDegraded`
- `ResourceStatusDisconnected`
- `ResourceStatusEvent`
- `ResourceStatusHealthy`
- `ResourceStatusReadonly`
- `ResourceStatusRecovering`
- `ResourceStatusUnknown`
- `RetryableError`
- `RuntimeContext`
- `RuntimeContextImpl`
- `Serve`
- `ServeConnector`
- `ServeFactory`
- `ServeOptions`
- `SetTraceID`
- `SkippableError`
- `StartArgs`
- `StartReply`
- `StateGetReply`
- `StateKeyArgs`
- `StateManager`
- `StateSetArgs`
- `StateSetWithTTLArgs`
- `StateStore`
- `StatefulRuntimeContext`
- `StopReply`
- `StreamAckRequest`
- `StreamCloseRequest`
- `StreamConsumer`
- `StreamDeliveryAtLeastOnce`
- `StreamDeliveryBestEffort`
- `StreamDeliverySemantics`
- `StreamGrantCreditsRequest`
- `StreamMessage`
- `StreamNackRequest`
- `StreamOpenRequest`
- `StreamOpenResponse`
- `StreamReceiveRequest`
- `StreamReceiveResponse`
- `WarmupPolicy`
- `WarmupPolicyAlwaysOn`
- `WarmupPolicyLazy`
- `WarmupPolicyPreload`

package contractv1

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/transport"
	"github.com/punk-one/iot-rule-plugin-sdk/schema"
)

type Snapshot struct {
	Version   string                  `json:"version"`
	Functions map[string]string       `json:"functions"`
	Types     map[string]TypeContract `json:"types"`
	Constants map[string]string       `json:"constants"`
	Wire      WireContract            `json:"wire"`
}

type TypeContract struct {
	Kind           string   `json:"kind"`
	Underlying     string   `json:"underlying,omitempty"`
	Fields         []string `json:"fields,omitempty"`
	Methods        []string `json:"methods,omitempty"`
	PointerMethods []string `json:"pointer_methods,omitempty"`
}

type WireContract struct {
	ProtocolVersion  uint     `json:"protocol_version"`
	MagicCookieKey   string   `json:"magic_cookie_key"`
	MagicCookieValue string   `json:"magic_cookie_value"`
	DispenseKeys     []string `json:"dispense_keys"`
	RPCMethods       []string `json:"rpc_methods"`
	ErrorFixtures    []string `json:"error_fixtures"`
}

func Build(moduleRoot string) (Snapshot, error) {
	types := publicTypes()
	names := make(map[reflect.Type]string, len(types))
	keys := sortedKeys(types)
	for _, name := range keys {
		if _, exists := names[types[name]]; !exists {
			names[types[name]] = name
		}
	}

	typeContracts := make(map[string]TypeContract, len(types))
	for _, name := range keys {
		typeContracts[name] = describeType(types[name], names)
	}

	methods, dispense, err := scanWireLiterals(moduleRoot)
	if err != nil {
		return Snapshot{}, err
	}
	handshake := sdk.DefaultHandshakeConfig()
	return Snapshot{
		Version:   "v1",
		Functions: publicFunctions(names),
		Types:     typeContracts,
		Constants: publicConstants(),
		Wire: WireContract{
			ProtocolVersion:  handshake.ProtocolVersion,
			MagicCookieKey:   handshake.MagicCookieKey,
			MagicCookieValue: handshake.MagicCookieValue,
			DispenseKeys:     dispense,
			RPCMethods:       methods,
			ErrorFixtures: []string{
				core.EncodePluginError(core.NewRetryableError("fixture")),
				core.EncodePluginError(core.ErrNotImplemented),
				core.EncodePluginError(core.ErrBatchNotSupported),
			},
		},
	}, nil
}

func publicTypes() map[string]reflect.Type {
	return map[string]reflect.Type{
		"AckArgs":                            reflect.TypeOf(sdk.AckArgs{}),
		"AckMessage":                         reflect.TypeOf(sdk.AckMessage{}),
		"BasePlugin":                         reflect.TypeOf(sdk.BasePlugin{}),
		"BuiltinRuntimeContext":              reflect.TypeOf((*sdk.BuiltinRuntimeContext)(nil)).Elem(),
		"ConnectorBindingSpec":               reflect.TypeOf(sdk.ConnectorBindingSpec{}),
		"ConnectorClient":                    reflect.TypeOf((*sdk.ConnectorClient)(nil)).Elem(),
		"ConnectorDescriptor":                reflect.TypeOf(sdk.ConnectorDescriptor{}),
		"ConnectorPlugin":                    reflect.TypeOf((*sdk.ConnectorPlugin)(nil)).Elem(),
		"ConnectorRPC":                       reflect.TypeOf(sdk.ConnectorRPC{}),
		"ConnectorRPCClient":                 reflect.TypeOf(sdk.ConnectorRPCClient{}),
		"ConnectorRPCServer":                 reflect.TypeOf(sdk.ConnectorRPCServer{}),
		"ConnectorRequest":                   reflect.TypeOf(sdk.ConnectorRequest{}),
		"ConnectorRequestKind":               reflect.TypeOf(sdk.ConnectorRequestKind("")),
		"ConnectorResource":                  reflect.TypeOf(sdk.ConnectorResource{}),
		"ConnectorResponse":                  reflect.TypeOf(sdk.ConnectorResponse{}),
		"DefaultLogger":                      reflect.TypeOf(sdk.DefaultLogger{}),
		"DefaultMetrics":                     reflect.TypeOf(sdk.DefaultMetrics{}),
		"EmitBatchArgs":                      reflect.TypeOf(sdk.EmitBatchArgs{}),
		"Emitter":                            reflect.TypeOf((*sdk.Emitter)(nil)).Elem(),
		"EmitterAdapter":                     reflect.TypeOf(sdk.EmitterAdapter{}),
		"EngineBridgeRuntimeContext":         reflect.TypeOf((*sdk.EngineBridgeRuntimeContext)(nil)).Elem(),
		"EngineRPC":                          reflect.TypeOf((*sdk.EngineRPC)(nil)).Elem(),
		"EngineRPCClient":                    reflect.TypeOf(sdk.EngineRPCClient{}),
		"ErrorKind":                          reflect.TypeOf(sdk.ErrorKind("")),
		"Event":                              reflect.TypeOf(sdk.Event{}),
		"Field":                              reflect.TypeOf(sdk.Field{}),
		"HealthHeartbeat":                    reflect.TypeOf(sdk.HealthHeartbeat{}),
		"HealthKind":                         reflect.TypeOf(sdk.HealthKind("")),
		"HealthOptions":                      reflect.TypeOf(sdk.HealthOptions{}),
		"HealthPolicy":                       reflect.TypeOf(sdk.HealthPolicy{}),
		"HealthReport":                       reflect.TypeOf(sdk.HealthReport{}),
		"HealthReporter":                     reflect.TypeOf((*sdk.HealthReporter)(nil)).Elem(),
		"HealthRuntimeStats":                 reflect.TypeOf(sdk.HealthRuntimeStats{}),
		"HealthStatus":                       reflect.TypeOf(sdk.HealthStatus("")),
		"InfoReply":                          reflect.TypeOf(sdk.InfoReply{}),
		"InitArgs":                           reflect.TypeOf(sdk.InitArgs{}),
		"InitReply":                          reflect.TypeOf(sdk.InitReply{}),
		"LogBatchArgs":                       reflect.TypeOf(sdk.LogBatchArgs{}),
		"LogLevel":                           reflect.TypeOf(sdk.LogLevel("")),
		"Logger":                             reflect.TypeOf((*sdk.Logger)(nil)).Elem(),
		"MetricArgs":                         reflect.TypeOf(sdk.MetricArgs{}),
		"Metrics":                            reflect.TypeOf((*sdk.Metrics)(nil)).Elem(),
		"Plugin":                             reflect.TypeOf((*sdk.Plugin)(nil)).Elem(),
		"PluginCapabilities":                 reflect.TypeOf(sdk.PluginCapabilities{}),
		"PluginConfig":                       reflect.TypeOf(sdk.PluginConfig{}),
		"PluginError":                        reflect.TypeOf(sdk.PluginError{}),
		"PluginInfo":                         reflect.TypeOf(sdk.PluginInfo{}),
		"PluginRPC":                          reflect.TypeOf(sdk.PluginRPC{}),
		"PluginRPCClient":                    reflect.TypeOf(sdk.PluginRPCClient{}),
		"PluginRPCServer":                    reflect.TypeOf(sdk.PluginRPCServer{}),
		"PluginType":                         reflect.TypeOf(sdk.PluginType("")),
		"ProviderPolicy":                     reflect.TypeOf(sdk.ProviderPolicy{}),
		"PublishAckArgs":                     reflect.TypeOf(sdk.PublishAckArgs{}),
		"QuotaPolicy":                        reflect.TypeOf(sdk.QuotaPolicy{}),
		"ReceiveEventArgs":                   reflect.TypeOf(sdk.ReceiveEventArgs{}),
		"ReceiveEventReply":                  reflect.TypeOf(sdk.ReceiveEventReply{}),
		"ReceiveEventsArgs":                  reflect.TypeOf(sdk.ReceiveEventsArgs{}),
		"ReceiveEventsReply":                 reflect.TypeOf(sdk.ReceiveEventsReply{}),
		"ReportHealthArgs":                   reflect.TypeOf(sdk.ReportHealthArgs{}),
		"ResourceStatus":                     reflect.TypeOf(sdk.ResourceStatus("")),
		"ResourceStatusEvent":                reflect.TypeOf(sdk.ResourceStatusEvent{}),
		"RuntimeContext":                     reflect.TypeOf((*sdk.RuntimeContext)(nil)).Elem(),
		"RuntimeContextImpl":                 reflect.TypeOf(sdk.RuntimeContextImpl{}),
		"ServeOptions":                       reflect.TypeOf(sdk.ServeOptions{}),
		"StartArgs":                          reflect.TypeOf(sdk.StartArgs{}),
		"StartReply":                         reflect.TypeOf(sdk.StartReply{}),
		"StateGetReply":                      reflect.TypeOf(sdk.StateGetReply{}),
		"StateKeyArgs":                       reflect.TypeOf(sdk.StateKeyArgs{}),
		"StateManager":                       reflect.TypeOf((*sdk.StateManager)(nil)).Elem(),
		"StateSetArgs":                       reflect.TypeOf(sdk.StateSetArgs{}),
		"StateSetWithTTLArgs":                reflect.TypeOf(sdk.StateSetWithTTLArgs{}),
		"StateStore":                         reflect.TypeOf((*sdk.StateStore)(nil)).Elem(),
		"StatefulRuntimeContext":             reflect.TypeOf((*sdk.StatefulRuntimeContext)(nil)).Elem(),
		"StopReply":                          reflect.TypeOf(sdk.StopReply{}),
		"StreamAckRequest":                   reflect.TypeOf(sdk.StreamAckRequest{}),
		"StreamCloseRequest":                 reflect.TypeOf(sdk.StreamCloseRequest{}),
		"StreamConsumer":                     reflect.TypeOf((*sdk.StreamConsumer)(nil)).Elem(),
		"StreamDeliverySemantics":            reflect.TypeOf(sdk.StreamDeliverySemantics("")),
		"StreamGrantCreditsRequest":          reflect.TypeOf(sdk.StreamGrantCreditsRequest{}),
		"StreamMessage":                      reflect.TypeOf(sdk.StreamMessage{}),
		"StreamNackRequest":                  reflect.TypeOf(sdk.StreamNackRequest{}),
		"StreamOpenRequest":                  reflect.TypeOf(sdk.StreamOpenRequest{}),
		"StreamOpenResponse":                 reflect.TypeOf(sdk.StreamOpenResponse{}),
		"StreamReceiveRequest":               reflect.TypeOf(sdk.StreamReceiveRequest{}),
		"StreamReceiveResponse":              reflect.TypeOf(sdk.StreamReceiveResponse{}),
		"WarmupPolicy":                       reflect.TypeOf(sdk.WarmupPolicy("")),
		"schema.FlowMetadata":                reflect.TypeOf(schema.FlowMetadata{}),
		"schema.GrayReleaseConfig":           reflect.TypeOf(schema.GrayReleaseConfig{}),
		"schema.NodeConfig":                  reflect.TypeOf(schema.NodeConfig{}),
		"schema.NodeMetadata":                reflect.TypeOf(schema.NodeMetadata{}),
		"schema.Position":                    reflect.TypeOf(schema.Position{}),
		"schema.RouteConfig":                 reflect.TypeOf(schema.RouteConfig{}),
		"schema.RuleFlow":                    reflect.TypeOf(schema.RuleFlow{}),
		"schema.RuleMessage":                 reflect.TypeOf(schema.RuleMessage{}),
		"schema.RuleSnapshot":                reflect.TypeOf(schema.RuleSnapshot{}),
		"schema.SourceConfig":                reflect.TypeOf(schema.SourceConfig{}),
		"wire.ConnectorAckStreamArgs":        reflect.TypeOf(transport.ConnectorAckStreamArgs{}),
		"wire.ConnectorAckStreamReply":       reflect.TypeOf(transport.ConnectorAckStreamReply{}),
		"wire.ConnectorCloseStreamArgs":      reflect.TypeOf(transport.ConnectorCloseStreamArgs{}),
		"wire.ConnectorCloseStreamReply":     reflect.TypeOf(transport.ConnectorCloseStreamReply{}),
		"wire.ConnectorCreateResourceArgs":   reflect.TypeOf(transport.ConnectorCreateResourceArgs{}),
		"wire.ConnectorCreateResourceReply":  reflect.TypeOf(transport.ConnectorCreateResourceReply{}),
		"wire.ConnectorDestroyResourceArgs":  reflect.TypeOf(transport.ConnectorDestroyResourceArgs{}),
		"wire.ConnectorDestroyResourceReply": reflect.TypeOf(transport.ConnectorDestroyResourceReply{}),
		"wire.ConnectorExecuteArgs":          reflect.TypeOf(transport.ConnectorExecuteArgs{}),
		"wire.ConnectorExecuteReply":         reflect.TypeOf(transport.ConnectorExecuteReply{}),
		"wire.ConnectorGrantCreditsArgs":     reflect.TypeOf(transport.ConnectorGrantCreditsArgs{}),
		"wire.ConnectorGrantCreditsReply":    reflect.TypeOf(transport.ConnectorGrantCreditsReply{}),
		"wire.ConnectorNackStreamArgs":       reflect.TypeOf(transport.ConnectorNackStreamArgs{}),
		"wire.ConnectorNackStreamReply":      reflect.TypeOf(transport.ConnectorNackStreamReply{}),
		"wire.ConnectorOpenStreamArgs":       reflect.TypeOf(transport.ConnectorOpenStreamArgs{}),
		"wire.ConnectorOpenStreamReply":      reflect.TypeOf(transport.ConnectorOpenStreamReply{}),
		"wire.ConnectorProbeArgs":            reflect.TypeOf(transport.ConnectorProbeArgs{}),
		"wire.ConnectorProbeReply":           reflect.TypeOf(transport.ConnectorProbeReply{}),
		"wire.ConnectorReceiveStreamArgs":    reflect.TypeOf(transport.ConnectorReceiveStreamArgs{}),
		"wire.ConnectorReceiveStreamReply":   reflect.TypeOf(transport.ConnectorReceiveStreamReply{}),
		"wire.CurrentResourceStatusArgs":     reflect.TypeOf(transport.CurrentResourceStatusArgs{}),
		"wire.CurrentResourceStatusReply":    reflect.TypeOf(transport.CurrentResourceStatusReply{}),
		"wire.NextResourceEventArgs":         reflect.TypeOf(transport.NextResourceEventArgs{}),
		"wire.NextResourceEventReply":        reflect.TypeOf(transport.NextResourceEventReply{}),
	}
}

func publicFunctions(names map[reflect.Type]string) map[string]string {
	values := map[string]interface{}{
		"DecodeEvent": sdk.DecodeEvent, "DefaultHandshakeConfig": sdk.DefaultHandshakeConfig,
		"DefaultHealthOptions": sdk.DefaultHealthOptions, "EdgeSubject": sdk.EdgeSubject,
		"EncodeEvent": sdk.EncodeEvent, "FatalError": sdk.FatalError,
		"FlushHealthReporter": sdk.FlushHealthReporter, "GetHealthRuntimeStats": sdk.GetHealthRuntimeStats,
		"GetTraceID": sdk.GetTraceID, "IsBatchNotSupported": sdk.IsBatchNotSupported,
		"IsFatal": sdk.IsFatal, "IsRetryable": sdk.IsRetryable, "IsSkippable": sdk.IsSkippable,
		"NewEngineRPCClient": sdk.NewEngineRPCClient, "NewEvent": sdk.NewEvent,
		"NewFatalError": sdk.NewFatalError, "NewRPCEmitterFromEngine": sdk.NewRPCEmitterFromEngine,
		"NewRPCLoggerFromEngine": sdk.NewRPCLoggerFromEngine, "NewRPCMetricsFromEngine": sdk.NewRPCMetricsFromEngine,
		"NewRetryableError": sdk.NewRetryableError, "NewSkippableError": sdk.NewSkippableError,
		"NewStateManager": sdk.NewStateManager, "RetryableError": sdk.RetryableError,
		"Serve": sdk.Serve, "ServeConnector": sdk.ServeConnector, "SetTraceID": sdk.SetTraceID,
		"SkippableError": sdk.SkippableError,
	}
	out := make(map[string]string, len(values))
	for name, value := range values {
		out[name] = canonicalType(reflect.TypeOf(value), names)
	}
	return out
}

func publicConstants() map[string]string {
	return map[string]string{
		"AckStatusSuccess": constant(sdk.AckStatusSuccess), "AckStatusFailed": constant(sdk.AckStatusFailed),
		"ErrorKindRetryable": constant(sdk.ErrorKindRetryable), "ErrorKindSkippable": constant(sdk.ErrorKindSkippable),
		"ErrorKindFatal": constant(sdk.ErrorKindFatal), "ErrorKindNotImplemented": constant(sdk.ErrorKindNotImplemented),
		"ErrorKindBatchUnsupported": constant(sdk.ErrorKindBatchUnsupported),
		"HealthUnknown":             constant(sdk.HealthUnknown), "HealthInitializing": constant(sdk.HealthInitializing),
		"HealthHealthy": constant(sdk.HealthHealthy), "HealthDegraded": constant(sdk.HealthDegraded),
		"HealthRecovering": constant(sdk.HealthRecovering), "HealthUnhealthy": constant(sdk.HealthUnhealthy),
		"HealthStopped": constant(sdk.HealthStopped), "HealthKindState": constant(sdk.HealthKindState),
		"HealthKindHeartbeat": constant(sdk.HealthKindHeartbeat),
		"PluginTypeSource":    constant(sdk.PluginTypeSource), "PluginTypeProcessor": constant(sdk.PluginTypeProcessor),
		"PluginTypeSink": constant(sdk.PluginTypeSink), "PluginTypeUtility": constant(sdk.PluginTypeUtility),
		"PluginTypeConnector": constant(sdk.PluginTypeConnector),
		"WarmupPolicyLazy":    constant(sdk.WarmupPolicyLazy), "WarmupPolicyPreload": constant(sdk.WarmupPolicyPreload),
		"WarmupPolicyAlwaysOn":                 constant(sdk.WarmupPolicyAlwaysOn),
		"ConnectorSessionPolicyPerOperation":   constant(sdk.ConnectorSessionPolicyPerOperation),
		"ConnectorSessionPolicyScheduledReuse": constant(sdk.ConnectorSessionPolicyScheduledReuse),
		"ConnectorSessionPolicyPersistent":     constant(sdk.ConnectorSessionPolicyPersistent),
		"ConnectorSessionPolicyPooled":         constant(sdk.ConnectorSessionPolicyPooled),
		"ConnectorSessionPolicyStateless":      constant(sdk.ConnectorSessionPolicyStateless),
		"ConnectorSessionPolicyIdleTimeout":    constant(sdk.ConnectorSessionPolicyIdleTimeout),
		"ResourceStatusUnknown":                constant(sdk.ResourceStatusUnknown), "ResourceStatusHealthy": constant(sdk.ResourceStatusHealthy),
		"ResourceStatusDegraded": constant(sdk.ResourceStatusDegraded), "ResourceStatusCongested": constant(sdk.ResourceStatusCongested),
		"ResourceStatusDisconnected": constant(sdk.ResourceStatusDisconnected), "ResourceStatusReadonly": constant(sdk.ResourceStatusReadonly),
		"ResourceStatusBlocked": constant(sdk.ResourceStatusBlocked), "ResourceStatusRecovering": constant(sdk.ResourceStatusRecovering),
		"ConnectorRequestRead": constant(sdk.ConnectorRequestRead), "ConnectorRequestWrite": constant(sdk.ConnectorRequestWrite),
		"ConnectorRequestBatchRead": constant(sdk.ConnectorRequestBatchRead), "ConnectorRequestBatchWrite": constant(sdk.ConnectorRequestBatchWrite),
		"ConnectorRequestMetadata": constant(sdk.ConnectorRequestMetadata),
		"StreamDeliveryBestEffort": constant(sdk.StreamDeliveryBestEffort), "StreamDeliveryAtLeastOnce": constant(sdk.StreamDeliveryAtLeastOnce),
		"LogLevelDebug": constant(sdk.LogLevelDebug), "LogLevelInfo": constant(sdk.LogLevelInfo),
		"LogLevelWarn": constant(sdk.LogLevelWarn), "LogLevelError": constant(sdk.LogLevelError),
		"EventMetadataTraceIDKey": constant(sdk.EventMetadataTraceIDKey),
	}
}

func constant(value interface{}) string { return fmt.Sprintf("%T:%v", value, value) }

func describeType(t reflect.Type, names map[reflect.Type]string) TypeContract {
	contract := TypeContract{Kind: t.Kind().String(), Underlying: canonicalUnderlying(t, names)}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			contract.Fields = append(contract.Fields, fmt.Sprintf("%s %s tag=%q anonymous=%t", field.Name, canonicalType(field.Type, names), string(field.Tag), field.Anonymous))
		}
	}
	contract.Methods = methodsOf(t, names)
	if t.Kind() != reflect.Interface && t.Kind() != reflect.Ptr {
		contract.PointerMethods = methodsOf(reflect.PointerTo(t), names)
	}
	return contract
}

func methodsOf(t reflect.Type, names map[reflect.Type]string) []string {
	out := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		out = append(out, method.Name+" "+canonicalType(method.Type, names))
	}
	sort.Strings(out)
	return out
}

func canonicalUnderlying(t reflect.Type, names map[reflect.Type]string) string {
	if t.Kind() == reflect.Struct || t.Kind() == reflect.Interface {
		return ""
	}
	return canonicalTypeUnaliased(t, names)
}

func canonicalType(t reflect.Type, names map[reflect.Type]string) string {
	if name, ok := names[t]; ok {
		return name
	}
	return canonicalTypeUnaliased(t, names)
}

func canonicalTypeUnaliased(t reflect.Type, names map[reflect.Type]string) string {
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + canonicalType(t.Elem(), names)
	case reflect.Slice:
		return "[]" + canonicalType(t.Elem(), names)
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), canonicalType(t.Elem(), names))
	case reflect.Map:
		return "map[" + canonicalType(t.Key(), names) + "]" + canonicalType(t.Elem(), names)
	case reflect.Chan:
		return t.ChanDir().String() + " chan " + canonicalType(t.Elem(), names)
	case reflect.Func:
		parts := make([]string, t.NumIn())
		for i := range parts {
			parts[i] = canonicalType(t.In(i), names)
		}
		results := make([]string, t.NumOut())
		for i := range results {
			results[i] = canonicalType(t.Out(i), names)
		}
		variadic := ""
		if t.IsVariadic() {
			variadic = " variadic"
		}
		return "func(" + strings.Join(parts, ",") + ")(" + strings.Join(results, ",") + ")" + variadic
	case reflect.Interface:
		return "interface"
	}
	if t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return t.String()
}

func scanWireLiterals(root string) ([]string, []string, error) {
	fset := token.NewFileSet()
	methods := map[string]struct{}{}
	dispense := map[string]struct{}{}
	paths := []string{root + "/internal/transport/engine_rpc.go", root + "/internal/transport/plugin_rpc.go", root + "/internal/transport/connector_rpc.go", root + "/internal/transport/serve.go"}
	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, nil, err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			if strings.HasPrefix(value, "Engine.") || strings.HasPrefix(value, "Plugin.") {
				methods[value] = struct{}{}
			}
			if path == paths[3] && (value == "plugin" || value == "connector") {
				dispense[value] = struct{}{}
			}
			return true
		})
	}
	return sortedSet(methods), sortedSet(dispense), nil
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func sortedKeys[V any](values map[string]V) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func Marshal(snapshot Snapshot) ([]byte, error) {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

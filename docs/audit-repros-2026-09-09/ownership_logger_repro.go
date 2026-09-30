//go:build ignore

// Run from SDK root: go run docs/audit-repros-2026-09-09/ownership_logger_repro.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/punk-one/iot-rule-plugin-sdk"
	rt "github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
	"runtime/pprof"
	"time"
)

func flusherCount() int {
	var b bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&b, 2)
	return bytes.Count(b.Bytes(), []byte("(*EngineRPCLogger).startFlusher"))
}
func main() {
	e := sdk.NewEvent(map[string]interface{}{"empty": []string{}}, nil)
	before, _ := json.Marshal(e.Payload)
	cloned := e.Clone()
	after, _ := json.Marshal(cloned.Payload)
	fmt.Printf("empty_slice_before=%s after=%s\n", before, after)
	m := map[string]interface{}{"n": 1}
	e = sdk.NewEvent(map[string]interface{}{"pointer": &m}, nil)
	cloned = e.Clone()
	(*cloned.Payload["pointer"].(*map[string]interface{}))["n"] = 2
	fmt.Printf("pointer_map_original_after_clone_mutation=%v\n", m["n"])
	first := flusherCount()
	for i := 0; i < 20; i++ {
		ctx := rt.NewPluginRuntimeContext(nil, "audit-rule", "audit-node", sdk.HealthOptions{}, nil, false)
		_ = ctx.(interface{ Close() error }).Close()
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("logger_flushers_retained_after_20_context_closes=%d\n", flusherCount()-first)
}

package sdk_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

func TestNewEventIDsAreConcurrentUniqueAndKeepV1Format(t *testing.T) {
	const workers = 32
	const perWorker = 500
	ids := make(chan string, workers*perWorker)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				ids <- sdk.NewEvent(nil, nil).ID
			}
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[string]struct{}, workers*perWorker)
	for id := range ids {
		if !strings.HasPrefix(id, "evt_") || strings.Trim(id[4:], "0123456789") != "" {
			t.Fatalf("event ID no longer matches evt_<digits>: %q", id)
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate event ID: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestEventCloneProducesIndependentMaps(t *testing.T) {
	event := sdk.NewEvent(map[string]interface{}{
		"device": map[string]interface{}{
			"id": "plc-1",
		},
		"values": []interface{}{1, "x"},
	}, map[string]string{
		"source": "s7",
	})
	sdk.SetTraceID(&event, "trace-1")

	cloned := event.Clone()

	cloned.Payload["device"].(map[string]interface{})["id"] = "plc-2"
	cloned.Payload["values"].([]interface{})[0] = 2
	cloned.Metadata["source"] = "modbus"
	sdk.SetTraceID(&cloned, "trace-2")

	if event.Payload["device"].(map[string]interface{})["id"] != "plc-1" {
		t.Fatalf("expected original nested payload map to stay unchanged")
	}
	if event.Payload["values"].([]interface{})[0] != 1 {
		t.Fatalf("expected original payload slice to stay unchanged")
	}
	if event.Metadata["source"] != "s7" {
		t.Fatalf("expected original metadata to stay unchanged")
	}
	if sdk.GetTraceID(event) != "trace-1" {
		t.Fatalf("expected original trace id to stay unchanged")
	}
}

func TestEventCloneCopiesMutablePayloadContainers(t *testing.T) {
	raw := json.RawMessage(`{"value":1}`)
	event := sdk.NewEvent(map[string]interface{}{
		"bytes":   []byte{1, 2, 3},
		"raw":     raw,
		"strings": []string{"a", "b"},
		"ints":    []int{1, 2},
		"floats":  []float64{1.5, 2.5},
		"labels":  map[string]string{"site": "north"},
	}, nil)

	cloned := event.Clone()
	cloned.Payload["bytes"].([]byte)[0] = 9
	cloned.Payload["raw"].(json.RawMessage)[0] = '['
	cloned.Payload["strings"].([]string)[0] = "changed"
	cloned.Payload["ints"].([]int)[0] = 9
	cloned.Payload["floats"].([]float64)[0] = 9.5
	cloned.Payload["labels"].(map[string]string)["site"] = "south"

	if event.Payload["bytes"].([]byte)[0] != 1 {
		t.Fatal("original []byte was mutated")
	}
	if string(event.Payload["raw"].(json.RawMessage)) != `{"value":1}` {
		t.Fatal("original json.RawMessage was mutated")
	}
	if event.Payload["strings"].([]string)[0] != "a" {
		t.Fatal("original []string was mutated")
	}
	if event.Payload["ints"].([]int)[0] != 1 {
		t.Fatal("original []int was mutated")
	}
	if event.Payload["floats"].([]float64)[0] != 1.5 {
		t.Fatal("original []float64 was mutated")
	}
	if event.Payload["labels"].(map[string]string)["site"] != "north" {
		t.Fatal("original map[string]string was mutated")
	}
}

func TestEventCloneCopiesTypedNestedContainers(t *testing.T) {
	event := sdk.NewEvent(map[string]interface{}{
		"records": []map[string]interface{}{{
			"values": []int16{1, 2},
		}},
	}, nil)

	cloned := event.Clone()
	records := cloned.Payload["records"].([]map[string]interface{})
	records[0]["new"] = true
	records[0]["values"].([]int16)[0] = 9

	original := event.Payload["records"].([]map[string]interface{})
	if _, ok := original[0]["new"]; ok {
		t.Fatal("original typed nested map was mutated")
	}
	if original[0]["values"].([]int16)[0] != 1 {
		t.Fatal("original uncommon typed slice was mutated")
	}
}

func TestEventClonePreservesNilTypedSlices(t *testing.T) {
	var values []string
	event := sdk.NewEvent(map[string]interface{}{"values": values}, nil)

	cloned := event.Clone()
	if cloned.Payload["values"].([]string) != nil {
		t.Fatal("expected nil typed slice to remain nil")
	}
}

func TestEventClonePreservesNonNilEmptySlices(t *testing.T) {
	event := sdk.NewEvent(map[string]interface{}{
		"values": []string{},
		"bytes":  []byte{},
		"raw":    json.RawMessage{},
	}, nil)

	cloned := event.Clone()
	if cloned.Payload["values"].([]string) == nil {
		t.Fatal("expected non-nil empty []string to remain non-nil")
	}
	if cloned.Payload["bytes"].([]byte) == nil {
		t.Fatal("expected non-nil empty []byte to remain non-nil")
	}
	if cloned.Payload["raw"].(json.RawMessage) == nil {
		t.Fatal("expected non-nil empty json.RawMessage to remain non-nil")
	}
}

func TestEventCloneCopiesPointerBackedTypedPayload(t *testing.T) {
	values := map[string]interface{}{
		"nested": []int{1, 2},
	}
	event := sdk.NewEvent(map[string]interface{}{"values": &values}, nil)

	cloned := event.Clone()
	clonedValues := cloned.Payload["values"].(*map[string]interface{})
	(*clonedValues)["nested"].([]int)[0] = 9
	(*clonedValues)["new"] = true

	if values["nested"].([]int)[0] != 1 {
		t.Fatal("original pointer-backed nested slice was mutated")
	}
	if _, exists := values["new"]; exists {
		t.Fatal("original pointer-backed map was mutated")
	}
}

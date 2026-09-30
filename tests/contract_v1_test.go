package sdk_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/contractv1"
)

func TestV1ContractSnapshot(t *testing.T) {
	snapshot, err := contractv1.Build("..")
	if err != nil {
		t.Fatal(err)
	}
	got, err := contractv1.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../contract_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("v1 API or wire contract changed; run contract generator only after compatibility review")
	}
}

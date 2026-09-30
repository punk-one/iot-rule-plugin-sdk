package main

import (
	"os"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/contractv1"
)

func main() {
	snapshot, err := contractv1.Build(".")
	if err != nil {
		panic(err)
	}
	data, err := contractv1.Marshal(snapshot)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("contract_v1.json", data, 0o644); err != nil {
		panic(err)
	}
}

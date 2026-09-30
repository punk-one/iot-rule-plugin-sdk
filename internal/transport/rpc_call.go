package transport

import (
	"fmt"
	"net/rpc"
	"time"
)

const (
	defaultRPCTimeout   = 30 * time.Second
	telemetryRPCTimeout = 5 * time.Second
	rpcCloseWait        = time.Second
)

func callWithTimeout(client *rpc.Client, method string, args, reply interface{}, timeout time.Duration) error {
	if client == nil {
		return fmt.Errorf("RPC client is nil: %s", method)
	}
	if timeout <= 0 {
		return fmt.Errorf("RPC timeout: %s", method)
	}

	done := make(chan error, 1)
	go func() { done <- client.Call(method, args, reply) }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		// net/rpc cannot cancel an individual request. Closing the connection is
		// the only reliable way to release the pending call and prevents one
		// leaked goroutine per timeout. The host must replace this client.
		_ = client.Close()
		select {
		case <-done:
		case <-time.After(rpcCloseWait):
		}
		return fmt.Errorf("RPC timeout: %s", method)
	}
}

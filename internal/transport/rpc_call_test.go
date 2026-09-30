package transport

import (
	"net"
	"net/rpc"
	"strings"
	"testing"
	"time"
)

type blockingRPCService struct {
	started chan struct{}
	release chan struct{}
}

func (s *blockingRPCService) Block(_ struct{}, _ *struct{}) error {
	close(s.started)
	<-s.release
	return nil
}

func TestCallWithTimeoutClosesConnection(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	server := rpc.NewServer()
	service := &blockingRPCService{started: make(chan struct{}), release: make(chan struct{})}
	if err := server.RegisterName("Blocking", service); err != nil {
		t.Fatal(err)
	}
	go server.ServeConn(serverConn)
	client := rpc.NewClient(clientConn)

	result := make(chan error, 1)
	go func() {
		result <- callWithTimeout(client, "Blocking.Block", struct{}{}, &struct{}{}, 25*time.Millisecond)
	}()
	select {
	case <-service.started:
	case <-time.After(time.Second):
		t.Fatal("RPC did not reach server")
	}
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "RPC timeout: Blocking.Block") {
			t.Fatalf("unexpected timeout error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out call did not return")
	}
	close(service.release)
	var reply struct{}
	if err := client.Call("Blocking.Block", struct{}{}, &reply); err == nil {
		t.Fatal("expected timed-out client connection to be closed")
	}
}

func TestCallWithTimeoutRejectsNilClient(t *testing.T) {
	if err := callWithTimeout(nil, "Plugin.Info", struct{}{}, &struct{}{}, time.Second); err == nil {
		t.Fatal("expected nil client error")
	}
}

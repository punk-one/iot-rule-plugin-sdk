package healthruntime

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type recordingSender struct {
	mu        sync.Mutex
	reports   []ReportArgs
	failures  int
	started   chan struct{}
	release   chan struct{}
	startOnce sync.Once
}

func (s *recordingSender) SendHealth(args ReportArgs) error {
	if s.started != nil {
		s.startOnce.Do(func() { close(s.started) })
	}
	if s.release != nil {
		<-s.release
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failures > 0 {
		s.failures--
		return errors.New("temporary sender failure")
	}
	s.reports = append(s.reports, args)
	return nil
}

func (s *recordingSender) snapshot() []ReportArgs {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ReportArgs(nil), s.reports...)
}

func TestLatestStateSurvivesQueueOverflow(t *testing.T) {
	sender := &recordingSender{started: make(chan struct{}), release: make(chan struct{})}
	reporter := NewAsync(sender, Options{HeartbeatInterval: time.Hour, MaxSilencePeriod: time.Hour, QueueCapacity: 1})
	if err := reporter.Healthy("initial", nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sender.started:
	case <-time.After(time.Second):
		t.Fatal("sender did not start")
	}
	for _, status := range []Status{StatusDegraded, StatusRecovering, StatusUnhealthy, StatusStopped} {
		if err := reporter.Report(Report{Status: status, Message: string(status)}); err != nil {
			t.Fatal(err)
		}
	}
	close(sender.release)
	if err := Flush(reporter, 2*time.Second); err != nil {
		t.Fatalf("flush failed: %v", err)
	}
	if err := reporter.(interface{ Close() error }).Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	reports := sender.snapshot()
	if len(reports) < 2 || reports[len(reports)-1].Status != StatusStopped {
		t.Fatalf("latest accepted state was not delivered: %#v", reports)
	}
	if stats := Snapshot(reporter); stats.CoalescedStates == 0 {
		t.Fatal("expected overflow coalescing to be observable")
	}
}

func TestSenderFailureRetriesAndUpdatesStats(t *testing.T) {
	sender := &recordingSender{failures: 2}
	reporter := NewAsync(sender, Options{HeartbeatInterval: time.Hour, QueueCapacity: 1})
	if err := reporter.Unhealthy("DB_DOWN", "database unavailable", nil); err != nil {
		t.Fatal(err)
	}
	if err := Flush(reporter, 3*time.Second); err != nil {
		t.Fatalf("flush failed: %v", err)
	}
	if err := reporter.(interface{ Close() error }).Close(); err != nil {
		t.Fatal(err)
	}
	if got := len(sender.snapshot()); got != 1 {
		t.Fatalf("expected one successful delivery, got %d", got)
	}
	if got := Snapshot(reporter).SenderErrors; got != 2 {
		t.Fatalf("expected two recorded sender errors, got %d", got)
	}
}

func TestCloseRejectsNewReportsAfterDrain(t *testing.T) {
	sender := &recordingSender{}
	reporter := NewAsync(sender, Options{HeartbeatInterval: time.Hour, QueueCapacity: 1})
	if err := reporter.Healthy("ready", nil); err != nil {
		t.Fatal(err)
	}
	if err := reporter.(interface{ Close() error }).Close(); err != nil {
		t.Fatal(err)
	}
	if err := reporter.Unhealthy("LATE", "late", nil); !errors.Is(err, ErrReporterClosed) {
		t.Fatalf("expected ErrReporterClosed, got %v", err)
	}
	if got := len(sender.snapshot()); got != 1 {
		t.Fatalf("close did not drain accepted report: %d", got)
	}
}

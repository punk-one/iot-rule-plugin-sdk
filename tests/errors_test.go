package sdk_test

import (
	"errors"
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

func TestStructuredPluginErrors(t *testing.T) {
	retryable := sdk.NewRetryableError("temporary")
	if !sdk.IsRetryable(retryable) {
		t.Fatalf("expected retryable error")
	}
	if sdk.IsFatal(retryable) {
		t.Fatalf("retryable error must not be fatal")
	}

	fatal := sdk.FatalError(errors.New("boom"))
	if !sdk.IsFatal(fatal) {
		t.Fatalf("expected fatal wrapped error")
	}
}

func TestBatchUnsupportedErrorIdentity(t *testing.T) {
	err := sdk.ErrBatchNotSupported
	if !sdk.IsBatchNotSupported(err) {
		t.Fatalf("expected batch unsupported")
	}

	wrapped := sdk.FatalError(err)
	if !sdk.IsBatchNotSupported(wrapped) {
		t.Fatalf("wrapped error should preserve underlying batch unsupported sentinel")
	}
	if !sdk.IsFatal(wrapped) {
		t.Fatalf("wrapped error should still be fatal")
	}

	if !errors.Is(sdk.ErrBatchNotSupported, sdk.ErrBatchNotSupported) {
		t.Fatalf("expected errors.Is to work on predefined batch unsupported error")
	}
}

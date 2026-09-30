package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ErrorKind string

const (
	ErrorKindRetryable        ErrorKind = "retryable"
	ErrorKindSkippable        ErrorKind = "skippable"
	ErrorKindFatal            ErrorKind = "fatal"
	ErrorKindNotImplemented   ErrorKind = "not_implemented"
	ErrorKindBatchUnsupported ErrorKind = "batch_unsupported"
)

const pluginErrorWirePrefix = "PUNK_PLUGIN_ERROR:"

type PluginError struct {
	Kind    ErrorKind `json:"kind"`
	Message string    `json:"message"`
	cause   error
}

func (e *PluginError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *PluginError) Unwrap() error { return e.cause }

func (e *PluginError) Is(target error) bool {
	other, ok := target.(*PluginError)
	if !ok || e == nil || other == nil {
		return false
	}
	return e.Kind != "" && e.Kind == other.Kind
}

var (
	ErrNotImplemented    = &PluginError{Kind: ErrorKindNotImplemented, Message: "method not implemented"}
	ErrBatchNotSupported = &PluginError{Kind: ErrorKindBatchUnsupported, Message: "batch processing not supported"}
)

func NewRetryableError(msg string) error { return &PluginError{Kind: ErrorKindRetryable, Message: msg} }
func NewSkippableError(msg string) error { return &PluginError{Kind: ErrorKindSkippable, Message: msg} }
func NewFatalError(msg string) error     { return &PluginError{Kind: ErrorKindFatal, Message: msg} }

func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &PluginError{Kind: ErrorKindRetryable, Message: err.Error(), cause: err}
}

func SkippableError(err error) error {
	if err == nil {
		return nil
	}
	return &PluginError{Kind: ErrorKindSkippable, Message: err.Error(), cause: err}
}

func FatalError(err error) error {
	if err == nil {
		return nil
	}
	return &PluginError{Kind: ErrorKindFatal, Message: err.Error(), cause: err}
}

func IsRetryable(err error) bool { return errorKindIs(err, ErrorKindRetryable) }
func IsSkippable(err error) bool { return errorKindIs(err, ErrorKindSkippable) }
func IsFatal(err error) bool     { return errorKindIs(err, ErrorKindFatal) }

func IsBatchNotSupported(err error) bool {
	return errors.Is(err, ErrBatchNotSupported) || errorKindIs(err, ErrorKindBatchUnsupported)
}

func errorKindIs(err error, expected ErrorKind) bool {
	var pluginErr *PluginError
	if !errors.As(err, &pluginErr) {
		return false
	}
	return pluginErr.Kind == expected
}

func EncodePluginError(err error) string {
	if err == nil {
		return ""
	}

	var pluginErr *PluginError
	if !errors.As(err, &pluginErr) {
		return err.Error()
	}

	payload, marshalErr := json.Marshal(struct {
		Kind    ErrorKind `json:"kind"`
		Message string    `json:"message"`
	}{
		Kind:    pluginErr.Kind,
		Message: pluginErr.Message,
	})
	if marshalErr != nil {
		return pluginErr.Error()
	}
	return pluginErrorWirePrefix + string(payload)
}

func DecodePluginError(raw string) error {
	if raw == "" {
		return nil
	}
	if !strings.HasPrefix(raw, pluginErrorWirePrefix) {
		return errors.New(raw)
	}

	var payload struct {
		Kind    ErrorKind `json:"kind"`
		Message string    `json:"message"`
	}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(raw, pluginErrorWirePrefix)), &payload); err != nil {
		return fmt.Errorf("malformed plugin error payload: %w", err)
	}
	return &PluginError{Kind: payload.Kind, Message: payload.Message}
}

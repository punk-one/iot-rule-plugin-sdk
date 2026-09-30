package sdk

import "github.com/punk-one/iot-rule-plugin-sdk/internal/core"

type ErrorKind = core.ErrorKind

const (
	ErrorKindRetryable        = core.ErrorKindRetryable
	ErrorKindSkippable        = core.ErrorKindSkippable
	ErrorKindFatal            = core.ErrorKindFatal
	ErrorKindNotImplemented   = core.ErrorKindNotImplemented
	ErrorKindBatchUnsupported = core.ErrorKindBatchUnsupported
)

type PluginError = core.PluginError

var (
	ErrNotImplemented    = core.ErrNotImplemented
	ErrBatchNotSupported = core.ErrBatchNotSupported
)

func NewRetryableError(msg string) error { return core.NewRetryableError(msg) }
func NewSkippableError(msg string) error { return core.NewSkippableError(msg) }
func NewFatalError(msg string) error     { return core.NewFatalError(msg) }

func RetryableError(err error) error { return core.RetryableError(err) }
func SkippableError(err error) error { return core.SkippableError(err) }
func FatalError(err error) error     { return core.FatalError(err) }

func IsRetryable(err error) bool { return core.IsRetryable(err) }
func IsSkippable(err error) bool { return core.IsSkippable(err) }
func IsFatal(err error) bool     { return core.IsFatal(err) }

func IsBatchNotSupported(err error) bool {
	return core.IsBatchNotSupported(err)
}

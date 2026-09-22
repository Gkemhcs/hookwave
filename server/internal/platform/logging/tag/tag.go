// Package tag provides typed log-field helpers, used by callers that only
// need field construction and don't want to depend on the rest of the
// logging package.
package tag

import (
	"time"

	"go.uber.org/zap"
)

// NewTag builds a zap.Field for key/value, picking the right typed
// constructor based on value's concrete type (falling back to zap.Any).
func NewTag(key string, value any) zap.Field {
	switch value := value.(type) {
	case int:
		return zap.Int(key, value)
	case int32:
		return zap.Int32(key, value)
	case int64:
		return zap.Int64(key, value)
	case string:
		return zap.String(key, value)
	case time.Duration:
		return zap.Duration(key, value)
	default:
		return zap.Any(key, value)
	}

}

// NewTagError builds a zap.Field for an error value.
func NewTagError(err error) zap.Field {
	return zap.Error(err)
}

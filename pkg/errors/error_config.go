package errors

import (
	"context"
)

// errorConfigContextKey is the context key for storing the error config.
var errorConfigContextKey int

// WithErrorConfig adds an ErrorConfig to the context.
func WithErrorConfig(parent context.Context, config *ErrorConfig) context.Context {
	if config == nil {
		return parent
	}
	return context.WithValue(parent, &errorConfigContextKey, config)
}

// ErrorConfigFromContext returns the ErrorConfig from the context, or nil if not set.
func ErrorConfigFromContext(ctx context.Context) *ErrorConfig {
	if ctx == nil {
		return nil
	}
	config := ctx.Value(&errorConfigContextKey)
	if config != nil {
		return config.(*ErrorConfig)
	}
	return nil
}

// ErrorCallback is a function that can modify an error before it is returned.
type ErrorCallback func(ctx context.Context, err ValidationError) ValidationError

// ErrorConfig holds error customization options that can be set on RuleSets.
type ErrorConfig struct {
	Short    string         // Custom short description
	Long     string         // Custom long description
	DocsURI  string         // Custom documentation URI
	TraceURI string         // Custom trace/debug URI
	Code     *ErrorCode     // Custom error code (nil means use default)
	Meta     map[string]any // Additional metadata
	Callback ErrorCallback  // Custom error callback
}

// ErrorConfigurable is a generic interface for types that support error customization.
// The Self type parameter enables method chaining by returning the concrete type.
// Use this with F-bounded polymorphism: `func Foo[T any, RS ErrorConfigurable[T, RS]](rs RS) RS`
type ErrorConfigurable[T any, Self any] interface {
	WithErrorMessage(short, long string) Self
	WithDocsURI(uri string) Self
	WithTraceURI(uri string) Self
	WithErrorCode(code ErrorCode) Self
	WithErrorMeta(key string, value any) Self
	WithErrorCallback(fn ErrorCallback) Self
}

// WithErrorMessage returns a new ErrorConfig with the given short and long messages merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithErrorMessage(short, long string) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{Short: short, Long: long})
}

// WithDocsURI returns a new ErrorConfig with the given documentation URI merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithDocsURI(uri string) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{DocsURI: uri})
}

// WithTraceURI returns a new ErrorConfig with the given trace URI merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithTraceURI(uri string) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{TraceURI: uri})
}

// WithCode returns a new ErrorConfig with the given error code merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithCode(code ErrorCode) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{Code: &code})
}

// WithMeta returns a new ErrorConfig with the given metadata key-value pair merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithMeta(key string, value any) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{Meta: map[string]any{key: value}})
}

// WithCallback returns a new ErrorConfig with the given callback merged.
// This method is nil-receiver safe.
func (c *ErrorConfig) WithCallback(fn ErrorCallback) *ErrorConfig {
	return mergeErrorConfig(c, &ErrorConfig{Callback: fn})
}

// mergeErrorConfig merges error configs from parent to child, with child taking precedence.
func mergeErrorConfig(parent, child *ErrorConfig) *ErrorConfig {
	if child == nil && parent == nil {
		return nil
	}
	if child == nil {
		return parent
	}
	if parent == nil {
		return child
	}

	merged := &ErrorConfig{}

	// Child values take precedence
	if child.Short != "" {
		merged.Short = child.Short
	} else {
		merged.Short = parent.Short
	}

	if child.Long != "" {
		merged.Long = child.Long
	} else {
		merged.Long = parent.Long
	}

	if child.DocsURI != "" {
		merged.DocsURI = child.DocsURI
	} else {
		merged.DocsURI = parent.DocsURI
	}

	if child.TraceURI != "" {
		merged.TraceURI = child.TraceURI
	} else {
		merged.TraceURI = parent.TraceURI
	}

	if child.Code != nil {
		merged.Code = child.Code
	} else {
		merged.Code = parent.Code
	}

	// Merge metadata maps
	if parent.Meta != nil || child.Meta != nil {
		merged.Meta = make(map[string]any)
		for k, v := range parent.Meta {
			merged.Meta[k] = v
		}
		for k, v := range child.Meta {
			merged.Meta[k] = v
		}
	}

	// Use child callback, or parent if child is nil
	if child.Callback != nil {
		merged.Callback = child.Callback
	} else {
		merged.Callback = parent.Callback
	}

	return merged
}

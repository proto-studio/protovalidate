<picture alt="Proto://Validate">
  <source media="(prefers-color-scheme: dark)" srcset="./docs/ProtoValidate_dark.svg">
  <img src="./docs/ProtoValidate_light.svg" width="800">
</picture>

[![Tests](https://github.com/proto-studio/protovalidate/actions/workflows/tests.yml/badge.svg)](https://github.com/proto-studio/protovalidate/actions/workflows/tests.yml)
[![License](https://img.shields.io/badge/License-BSD%203%20Clause-blue.svg)](https://github.com/proto-studio/protovalidate/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/proto.zip/studio/validate)
[![codecov](https://codecov.io/gh/proto-studio/protovalidate/graph/badge.svg?token=K6MR10DKYX)](https://codecov.io/gh/proto-studio/protovalidate)
[![Go Report Card](https://goreportcard.com/badge/proto.zip/studio/validate)](https://goreportcard.com/report/proto.zip/studio/validate)
[![Discord Chat](https://img.shields.io/badge/Discord-chat-blue?logo=Discord&logoColor=white)](https://proto.studio/social/discord)

ProtoValidate is a fluent data validation and normalization library for Go.

Project goals:

1. Human readable easy to understand and write validation rules.
2. Easily extensible / composable to support any data type and rules.
3. Detailed, actionable, and customizable errors.

Features:

- Type checking at compile time.
- Declarative rule syntax.
- Works on deeply nested objects and slices/arrays.
- Supports automatic correction and normalization of data.
- Easy to extend with custom validation rules.
- Easy to support additional data types.
- Structured error responses make it easy to format, display, and correct errors.
- Support for Internationalization (i18n) for error messages.
- Able to convert unstructured data (such as Json) to structured typed data.

Common use cases:

- API input validation.
- Command line flag validation.
- Unit testing.
- File validating.

Supported data types out of the box:
- `string`
- `int` / `int8` / `int16` / `int32` / `int64`
- `uint` / `uint8` / `uint16` / `uint32` / `uint64`
- `float32` / `float64`
- `struct` / `map` / `[]`
- `time.Time`
- Email addresses
- Domains

Easily customize to make support your own date types.

## Versioning

This package follows conventional Go versioning. Any version up to version 1.0.0 is considered "unstable" and the API may change.

We put a lot of thought into the design of this library and don't expect there to be many breaking changes. You are free to use this library in a production setting. However, keep an eye on the release notes as it will be rapidly changing.

## Getting Started
### Quick Start

```bash
go get proto.zip/studio/validate
```

Simple usage:

```go
package main

import (
        "fmt"
        "os"

        "proto.zip/studio/validate"
        "proto.zip/studio/validate/pkg/rules"
)

var ruleSet rules.RuleSet[string] = validate.String().
        WithMinLen(3).
        WithMaxLen(7)

// Try changing the string to see different results
func main() {
        str := "a"

        if _, err := ruleSet.Validate(str); err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
}
```

See the [examples](https://github.com/proto-studio/protovalidate/tree/main/examples) folder for more samples.

### Best Practices
Here are some best practices to help you get the most out of ProtoValidate:

- Break custom rules out into their own testable functions.
- If you need the whole object for your rule you can add it to the object `RuleSet` instead of the key.
- Don't redefine your rules sets every use. Rule sets are immutable so you can use/reuse the same instance in multiple threads.
- Impossible rules will often `panic` at runtime. Defining your rule sets at the top level or in module init will let you catch them at launch instead of later.

## Sponsors

- [ProtoStudio](https://proto.studio) - Build app backends fast without writing code.
- [Curioso Industries](https://curiosoindustries.com) - Expert product development and consulting services.

## Support

ProtoValidate is built for mission critical code. We want you to get all the support you need.

For community support join the [ProtoStudio Discord Community](https://proto.studio/social/discord). If you require commercial support please contact our premium support partner [Curioso Industries](https://curiosoindustries.com).

# Path Serialization Example

This example demonstrates how the same validation error can be serialized using different path formats.

## Running the Example

```bash
go run _examples/path_serialization/app.go
```

## What It Shows

The example creates validation errors with various path structures and displays how each path is represented in different serialization formats:

1. **Default Format**: The original format used by `Path()` method (`/segment1/segment2` or `0/1`)
2. **JSON Pointer (RFC 6901)**: Standard JSON Pointer format (`/segment1/segment2/0`)
3. **JSONPath**: JSONPath format (`$.segment1.segment2[0]`)
4. **Dot Notation**: Dot notation format (`segment1.segment2[0]`)

## Example Output

The program demonstrates:
- Simple nested paths (e.g., `users.profile.name`)
- Paths with array indices (e.g., `users[0].emails[1]`)
- Complex mixed paths (e.g., `data.items[2].metadata.tags[0]`)
- Paths starting with indices (e.g., `[5].value`)
- Single segment paths (e.g., `username`)

Each example shows how the same error path is represented across all four serialization formats.

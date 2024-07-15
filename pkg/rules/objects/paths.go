package objects

import (
	"fmt"
)

func toQuotedPath(key any) string {
	switch x := key.(type) {
	case string:
		return `"` + x + `"`
	}
	return fmt.Sprintf("%v", key)
}

func toPath(key any) string {
	return fmt.Sprintf("%v", key)
}

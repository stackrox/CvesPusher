package utils

import (
	"encoding/json"
	"fmt"
)

type stringerFunc func() string

func (f stringerFunc) String() string { return f() }

func AsJSON(x any) fmt.Stringer {
	return stringerFunc(func() string {
		bytes, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprintf("#ERROR(%v)", err)
		}
		return string(bytes)
	})
}

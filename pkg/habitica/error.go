package habitica

import "fmt"

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}

package errors

import (
	"fmt"
)

type FinderError struct {
	ret      string
	function string
	desc     string
}

func (fe *FinderError) Error() string {
	format := `An error caught in %s, %s[%s].`
	return fmt.Sprintf(format, fe.function, fe.desc, fe.ret)
}

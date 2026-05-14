package flatten

import (
	"github.com/MarkRosemaker/errpath"
	"github.com/MarkRosemaker/openapi"
)

func parameters(d *openapi.Document, ps openapi.Parameters) error {
	for name, p := range ps.ByIndex() {
		// NOTE: We are *not* calling parameterRef here,
		// because we are calling this function from Components,
		// where the parameter should already be.
		if err := parameter(d, p.Value); err != nil {
			return &errpath.ErrKey{Key: name, Err: err}
		}
	}

	return nil
}

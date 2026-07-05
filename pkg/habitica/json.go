package habitica

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"time"
)

func init() {
	jsonOpts = json.JoinOptions(
		json.RejectUnknownMembers(true),
		json.WithUnmarshalers(json.UnmarshalFromFunc(timeUnmarshal)),
	)
}

const altLayout = "Mon Jan 02 2006 15:04:05 MST-0700"

var layouts = []string{time.RFC3339, altLayout}

func timeUnmarshal(dec *jsontext.Decoder, t *time.Time) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return err
	}

	errs := make([]error, 0, len(layouts))
	for _, layout := range layouts {
		ts, err := time.Parse(layout, s)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		*t = ts
		return nil
	}

	return errors.Join(errs...)
}

package cassette

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"

	"github.com/MarkRosemaker/jsonutil"
)

var jsonOpts = json.JoinOptions(
	jsontext.Multiline(true),
	json.RejectUnknownMembers(true),
	json.WithMarshalers(json.MarshalToFunc(jsonutil.HTTPHeaderMarshal)),
	json.WithUnmarshalers(json.UnmarshalFromFunc(jsonutil.HTTPHeaderUnmarshal)),
)

func InteractionsReadFile(path string) (Interactions, error) {
	return jsonutil.ReadFile[Interactions](path, jsonOpts)
}

func InteractionsUnmarshal(data []byte) (Interactions, error) {
	out := Interactions{}
	return out, json.Unmarshal(data, &out, jsonOpts)
}

func InteractionsUnmarshalRead(r io.Reader) (Interactions, error) {
	out := Interactions{}
	return out, json.UnmarshalRead(r, &out, jsonOpts)
}

func (ias Interactions) WriteFile(path string) error {
	return jsonutil.WriteFile(path, ias, jsonOpts)
}

func (ias Interactions) MarshalWrite(w io.Writer) error {
	return json.MarshalWrite(w, ias, jsonOpts)
}

package cassette

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"

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

var mu sync.Mutex

// AddInteraction adds an interaction to the given path for debug purposes.
func AddInteraction(path string, ia Interaction) error {
	mu.Lock()
	defer mu.Unlock()

	ias, err := InteractionsReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading interactions file: %w", err)
	}

	if err := append(ias, ia).WriteFile(path); err != nil {
		return fmt.Errorf("writing interactions file: %w", err)
	}

	return nil
}

package habitica

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"

	"github.com/MarkRosemaker/cassette"
	"github.com/go-api-libs/api"
)

const (
	testAPIKey = "11111111-2222-3333-4444-555555555555"
	testUser   = "00000000-0000-0000-0000-000000000001"
	testTaskID = "2b774d70-ec8b-41c1-8967-eb6b13d962ba"
)

// recordingServer answers every request with status, contentType and body, and keeps the last request it got.
func recordingServer(t *testing.T, status int, contentType, body string) (*url.URL, *http.Request) {
	t.Helper()

	got := &http.Request{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = *r.Clone(t.Context())

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	return u, got
}

func TestNewClient(t *testing.T) {
	t.Run("no API key", func(t *testing.T) {
		t.Setenv("HABITICA_API_KEY", "")

		if _, err := NewClient(); err == nil || !strings.Contains(err.Error(), "HABITICA_API_KEY") {
			t.Fatalf("want error naming HABITICA_API_KEY, got: %v", err)
		}
	})

	t.Run("empty client parameter", func(t *testing.T) {
		t.Setenv("HABITICA_API_KEY", testAPIKey)

		if _, err := NewClient(WithClient("")); err == nil || !strings.Contains(err.Error(), "client parameter") {
			t.Fatalf("want client parameter error, got: %v", err)
		}
	})

	for _, tc := range []struct {
		name       string
		env        string
		opts       []ClientOption
		wantKey    string
		wantClient string
	}{
		{"key from environment", testAPIKey, nil, testAPIKey, "b0413351-405f-416f-8787-947ec1c85199-go-api-libs/habitica"},
		{"WithAPIKey overrides environment", "from-env", []ClientOption{WithAPIKey("from-option")}, "from-option", "b0413351-405f-416f-8787-947ec1c85199-go-api-libs/habitica"},
		{"WithAPIKey without environment", "", []ClientOption{WithAPIKey(testAPIKey)}, testAPIKey, "b0413351-405f-416f-8787-947ec1c85199-go-api-libs/habitica"},
		{"WithClient", testAPIKey, []ClientOption{WithClient("my-app")}, testAPIKey, "my-app"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HABITICA_API_KEY", tc.env)

			u, got := recordingServer(t, http.StatusTeapot, "application/json", "")

			c, err := NewClient(append(tc.opts, WithBaseURL(u))...)
			if err != nil {
				t.Fatal(err)
			}

			_, _ = c.Cron(t.Context(), CronParams{XAPIUser: uuid.MustParse(testUser)})

			if key := got.Header.Get("X-Api-Key"); key != tc.wantKey {
				t.Errorf("X-Api-Key: want %q, got %q", tc.wantKey, key)
			}

			if cl := got.Header.Get("X-Client"); cl != tc.wantClient {
				t.Errorf("X-Client: want %q, got %q", tc.wantClient, cl)
			}
		})
	}
}

func TestWithBaseURL_LeadingSlash(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	srv, got := recordingServer(t, http.StatusTeapot, "application/json", "")

	c, err := NewClient(WithBaseURL(&url.URL{Scheme: srv.Scheme, Host: srv.Host, Path: "api/v3"}))
	if err != nil {
		t.Fatal(err)
	}

	_, _ = c.GetUser(t.Context(), GetUserParams{})

	if got.URL.Path != "/api/v3/user" {
		t.Fatalf("want path /api/v3/user, got %q", got.URL.Path)
	}
}

func TestClient_Requests(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	user := uuid.MustParse(testUser)
	task := uuid.MustParse(testTaskID)
	target := uuid.MustParse("f7591dce-398a-4e12-ba23-18542f9f0bef")

	for _, tc := range []struct {
		name      string
		call      func(context.Context, *Client) error
		method    string
		path      string
		wantQuery url.Values
	}{
		{"GetUser", func(ctx context.Context, c *Client) error {
			_, err := c.GetUser(ctx, GetUserParams{XAPIUser: user})
			return err
		}, http.MethodGet, "/v3/user", url.Values{}},
		{"ListTasks all", func(ctx context.Context, c *Client) error {
			_, err := c.ListTasks(ctx, ListTasksParams{XAPIUser: user})
			return err
		}, http.MethodGet, "/v3/tasks/user", url.Values{}},
		{"ListTasks by type", func(ctx context.Context, c *Client) error {
			_, err := c.ListTasks(ctx, ListTasksParams{XAPIUser: user, Type: TypeCompletedTodos})
			return err
		}, http.MethodGet, "/v3/tasks/user", url.Values{"type": {"completedTodos"}}},
		{"GetTaskByID", func(ctx context.Context, c *Client) error {
			_, err := c.GetTaskByID(ctx, task, GetTaskByIDParams{XAPIUser: user})
			return err
		}, http.MethodGet, "/v3/tasks/" + testTaskID, url.Values{}},
		{"ScoreTask up", func(ctx context.Context, c *Client) error {
			_, err := c.ScoreTask(ctx, task, DirectionUp, ScoreTaskParams{XAPIUser: user})
			return err
		}, http.MethodPost, "/v3/tasks/" + testTaskID + "/score/up", url.Values{}},
		{"ScoreTask down", func(ctx context.Context, c *Client) error {
			_, err := c.ScoreTask(ctx, task, DirectionDown, ScoreTaskParams{XAPIUser: user})
			return err
		}, http.MethodPost, "/v3/tasks/" + testTaskID + "/score/down", url.Values{}},
		{"Cast on target", func(ctx context.Context, c *Client) error {
			_, err := c.Cast(ctx, SpellIDSmash, CastParams{XAPIUser: user, TargetID: target})
			return err
		}, http.MethodPost, "/v3/user/class/cast/smash", url.Values{"targetId": {target.String()}}},
		{"Cast without target", func(ctx context.Context, c *Client) error {
			_, err := c.Cast(ctx, SpellIDEarth, CastParams{XAPIUser: user})
			return err
		}, http.MethodPost, "/v3/user/class/cast/earth", url.Values{}},
		{"BuyHealthPotion", func(ctx context.Context, c *Client) error {
			_, err := c.BuyHealthPotion(ctx, BuyHealthPotionParams{XAPIUser: user})
			return err
		}, http.MethodPost, "/v3/user/buy-health-potion", url.Values{}},
		{"Cron", func(ctx context.Context, c *Client) error {
			_, err := c.Cron(ctx, CronParams{XAPIUser: user})
			return err
		}, http.MethodPost, "/v3/cron", url.Values{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, got := recordingServer(t, http.StatusTeapot, "application/json", "")

			c, err := NewClient(WithBaseURL(srv.JoinPath("v3")))
			if err != nil {
				t.Fatal(err)
			}

			if err := tc.call(t.Context(), c); !errors.Is(err, api.ErrUnknownStatusCode) {
				t.Fatalf("want %v, got: %v", api.ErrUnknownStatusCode, err)
			}

			if got.Method != tc.method {
				t.Errorf("method: want %s, got %s", tc.method, got.Method)
			}

			if got.URL.Path != tc.path {
				t.Errorf("path: want %s, got %s", tc.path, got.URL.Path)
			}

			if q := got.URL.Query(); q.Encode() != tc.wantQuery.Encode() {
				t.Errorf("query: want %q, got %q", tc.wantQuery.Encode(), q.Encode())
			}

			for name, want := range map[string]string{
				"X-Api-Key":  testAPIKey,
				"X-Api-User": testUser,
				"User-Agent": defaultUserAgent,
			} {
				if v := got.Header.Get(name); v != want {
					t.Errorf("%s: want %q, got %q", name, want, v)
				}
			}
		})
	}
}

func TestClient_Context(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	srv, _ := recordingServer(t, http.StatusOK, "application/json", "{}")

	c, err := NewClient(WithBaseURL(srv))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := c.GetUser(ctx, GetUserParams{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("want %v, got: %v", context.Canceled, err)
	}
}

// replayOne returns a client answering from the recorded interaction at index i of api/interactions.json.
func replayOne(t *testing.T, i int) *Client {
	t.Helper()
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	ias, err := cassette.InteractionsReadFile("../../api/interactions.json")
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(WithHTTPClient(&http.Client{Transport: cassette.Replay(ias[i : i+1])}))
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestClient_Responses(t *testing.T) {
	user := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	task := uuid.MustParse(testTaskID)

	t.Run("ListTasks", func(t *testing.T) {
		res, err := replayOne(t, 0).ListTasks(t.Context(), ListTasksParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		if !res.Success || len(res.Data) != 2 {
			t.Fatalf("want success with 2 tasks, got %v with %d", res.Success, len(res.Data))
		}

		if first := res.Data[0]; first.Text != "15 minute break" || first.Type != "reward" ||
			first.UnderscoreID != uuid.MustParse("8a9d461b-f5eb-4a16-97d3-c03380c422a3") {
			t.Errorf("first task: got %q (%s, %s)", first.Text, first.Type, first.UnderscoreID)
		}
	})

	t.Run("GetTaskByID", func(t *testing.T) {
		res, err := replayOne(t, 1).GetTaskByID(t.Context(), task, GetTaskByIDParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		if d := res.Data; d.UnderscoreID != task || d.Text != "API Trial" || d.Alias != "apiTrial" || d.Type != "habit" {
			t.Errorf("got task %s %q alias %q type %q", d.UnderscoreID, d.Text, d.Alias, d.Type)
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		res, err := replayOne(t, 2).GetUser(t.Context(), GetUserParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		if name := res.Data.Auth.Local.Username; name != "markrosemaker" {
			t.Errorf("username: got %q", name)
		}
	})

	t.Run("ScoreTask", func(t *testing.T) {
		res, err := replayOne(t, 3).ScoreTask(t.Context(), task, DirectionUp, ScoreTaskParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		d := res.Data
		if d.Delta != 1.0259567046270648 || d.UnderscoreTmp.Drop.Target != "Zombie" ||
			d.UnderscoreTmp.Quest.Collection == nil || *d.UnderscoreTmp.Quest.Collection != 1 {
			t.Errorf("got delta %v, drop target %q, quest %+v", d.Delta, d.UnderscoreTmp.Drop.Target, d.UnderscoreTmp.Quest)
		}
	})

	t.Run("ScoreTask not found", func(t *testing.T) {
		_, err := replayOne(t, 4).ScoreTask(t.Context(), uuid.MustParse("4f2d942d-67e8-4cc0-ac08-61962296660b"), DirectionUp, ScoreTaskParams{XAPIUser: user})
		assertServerError(t, err, http.StatusNotFound, "NotFound", "Not found.")
	})

	t.Run("Cast", func(t *testing.T) {
		res, err := replayOne(t, 6).Cast(t.Context(), SpellIDSmash, CastParams{
			XAPIUser: user, TargetID: uuid.MustParse("f7591dce-398a-4e12-ba23-18542f9f0bef"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if d := res.Data; d.Task.Text != "Early Bird #2 (sleeping at 10:30pm at the latest)" || d.User.Auth.Local.Username != "markrosemaker" {
			t.Errorf("got task %q, user %q", d.Task.Text, d.User.Auth.Local.Username)
		}
	})

	t.Run("Cast not enough mana", func(t *testing.T) {
		_, err := replayOne(t, 7).Cast(t.Context(), SpellIDSmash, CastParams{
			XAPIUser: user, TargetID: uuid.MustParse("f7591dce-398a-4e12-ba23-18542f9f0bef"),
		})
		assertServerError(t, err, http.StatusUnauthorized, "NotAuthorized", "Not enough mana.")
	})

	t.Run("BuyHealthPotion", func(t *testing.T) {
		res, err := replayOne(t, 8).BuyHealthPotion(t.Context(), BuyHealthPotionParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		if res.Message != "Bought Health Potion" || res.Data.Str != 32 || res.Data.Con != 22 {
			t.Errorf("got message %q, str %d, con %d", res.Message, res.Data.Str, res.Data.Con)
		}
	})

	t.Run("Cron", func(t *testing.T) {
		res, err := replayOne(t, 9).Cron(t.Context(), CronParams{XAPIUser: user})
		if err != nil {
			t.Fatal(err)
		}

		if !res.Success || res.UserV != 16098 || res.AppVersion != "5.49.0" {
			t.Errorf("got %+v", res)
		}
	})

	t.Run("ListTasks bad gateway", func(t *testing.T) {
		_, err := replayOne(t, 12).ListTasks(t.Context(), ListTasksParams{XAPIUser: user, Type: TypeDailys})

		body, ok := errors.AsType[*api.ErrorBody](err)
		if !ok {
			t.Fatalf("want *api.ErrorBody, got %T: %v", err, err)
		}

		if !strings.Contains(string(body.Body), "502 Server Error") {
			t.Errorf("body: %s", body.Body)
		}
	})
}

func TestClient_ListTasksBadRequest(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	srv, _ := recordingServer(t, http.StatusBadRequest, "application/json; charset=utf-8",
		`{"success":false,"error":"BadRequest","message":"Invalid type."}`)

	c, err := NewClient(WithBaseURL(srv))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.ListTasks(t.Context(), ListTasksParams{Type: "nonsense"})
	assertServerError(t, err, http.StatusBadRequest, "BadRequest", "Invalid type.")
}

func assertServerError(t *testing.T, err error, status int, code, message string) {
	t.Helper()

	apiErr, ok := errors.AsType[*api.Error](err)
	if !ok {
		t.Fatalf("want *api.Error, got %T: %v", err, err)
	}

	if apiErr.StatusCode() != status || !apiErr.IsCustom {
		t.Errorf("want custom error with status %d, got status %d custom %v", status, apiErr.StatusCode(), apiErr.IsCustom)
	}

	e, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}

	if e.Success || e.Err != code || e.Message != message {
		t.Errorf("got %+v", e)
	}

	if want := code + ": " + message; e.Error() != want {
		t.Errorf("Error(): want %q, got %q", want, e.Error())
	}
}

// cronWithExtra is a Cron response with a member the specification does not know.
const cronWithExtra = `{"success":true,"data":{},"notifications":[],"userV":7,"appVersion":"5.49.0","newMember":1}`

func TestClient_Decoding(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	srv, _ := recordingServer(t, http.StatusOK, "application/json", cronWithExtra)

	c, err := NewClient(WithBaseURL(srv))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("strict for the specified type", func(t *testing.T) {
		if _, err := c.Cron(t.Context(), CronParams{}); err == nil {
			t.Fatal("want decoding error for an unknown member")
		} else if _, ok := errors.AsType[*api.DecodingError](err); !ok {
			t.Fatalf("want *api.DecodingError, got %T: %v", err, err)
		}
	})

	t.Run("lenient for a custom result", func(t *testing.T) {
		res, err := c.CronWithResult[struct {
			UserV int `json:"userV"`
		}](t.Context(), CronParams{})
		if err != nil {
			t.Fatal(err)
		}

		if res.UserV != 7 {
			t.Errorf("userV: want 7, got %d", res.UserV)
		}
	})
}

func readInteractions(t *testing.T, path string) cassette.Interactions {
	t.Helper()

	ias, err := cassette.InteractionsReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return ias
}

func TestClient_Debug(t *testing.T) {
	t.Setenv("HABITICA_API_KEY", testAPIKey)

	t.Run("undecodable response is kept once and decoded leniently", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "rec", "interactions.json")
		srv, _ := recordingServer(t, http.StatusOK, "application/json", cronWithExtra)

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(path))
		if err != nil {
			t.Fatal(err)
		}

		for range 2 {
			res, err := c.Cron(t.Context(), CronParams{XAPIUser: uuid.MustParse(testUser)})
			if err != nil {
				t.Fatal(err)
			}

			if res.UserV != 7 || res.AppVersion != "5.49.0" {
				t.Errorf("got %+v", res)
			}
		}

		ias := readInteractions(t, path)
		if len(ias) != 1 {
			t.Fatalf("want 1 interaction kept, got %d", len(ias))
		}

		if ia := ias[0]; ia.Request.Method != http.MethodPost || ia.Request.URL.Path != "/cron" || ia.Response.StatusCode != http.StatusOK {
			t.Errorf("kept %s %s -> %d", ia.Request.Method, ia.Request.URL.Path, ia.Response.StatusCode)
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(raw), testAPIKey) {
			t.Error("the API key was written to the debug file")
		}
	})

	t.Run("decodable response is not kept", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "interactions.json")
		srv, _ := recordingServer(t, http.StatusOK, "application/json",
			`{"success":true,"data":{},"notifications":[],"userV":7,"appVersion":"5.49.0"}`)

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(path))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.Cron(t.Context(), CronParams{}); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("want no debug file, got: %v", err)
		}
	})

	t.Run("undecodable error response is decoded leniently", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "interactions.json")
		srv, _ := recordingServer(t, http.StatusNotFound, "application/json",
			`{"success":false,"error":"NotFound","message":"Not found.","extra":true}`)

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(path))
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.ScoreTask(t.Context(), uuid.MustParse(testTaskID), DirectionUp, ScoreTaskParams{})
		assertServerError(t, err, http.StatusNotFound, "NotFound", "Not found.")

		if n := len(readInteractions(t, path)); n != 1 {
			t.Fatalf("want 1 interaction kept, got %d", n)
		}
	})

	t.Run("invalid JSON is kept and still fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "interactions.json")
		srv, _ := recordingServer(t, http.StatusOK, "application/json", `{"success":`)

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(path))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.GetUser(t.Context(), GetUserParams{}); err == nil {
			t.Fatal("want error")
		} else if _, ok := errors.AsType[*api.DecodingError](err); !ok {
			t.Fatalf("want *api.DecodingError, got %T: %v", err, err)
		}

		if n := len(readInteractions(t, path)); n != 1 {
			t.Fatalf("want 1 interaction kept, got %d", n)
		}
	})

	for _, tc := range []struct {
		name        string
		status      int
		contentType string
		want        error
	}{
		{"unknown status code is kept", http.StatusTeapot, "application/json", api.ErrUnknownStatusCode},
		{"unknown content type is kept", http.StatusOK, "text/plain", api.ErrUnknownContentType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "interactions.json")
			srv, _ := recordingServer(t, tc.status, tc.contentType, "nope")

			c, err := NewClient(WithBaseURL(srv), WithDebugFile(path))
			if err != nil {
				t.Fatal(err)
			}

			if _, err := c.BuyHealthPotion(t.Context(), BuyHealthPotionParams{}); !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got: %v", tc.want, err)
			}

			ias := readInteractions(t, path)
			if len(ias) != 1 || ias[0].Response.StatusCode != tc.status {
				t.Fatalf("want the %d response kept, got %d interactions", tc.status, len(ias))
			}
		})
	}

	t.Run("failing to keep is joined to the error", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(blocker, nil, 0o600); err != nil {
			t.Fatal(err)
		}

		srv, _ := recordingServer(t, http.StatusTeapot, "application/json", "")

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(filepath.Join(blocker, "interactions.json")))
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.Cron(t.Context(), CronParams{})
		if !errors.Is(err, api.ErrUnknownStatusCode) {
			t.Fatalf("want %v, got: %v", api.ErrUnknownStatusCode, err)
		}

		if !strings.Contains(err.Error(), "interactions file") {
			t.Errorf("want the debug file error joined, got: %v", err)
		}
	})

	t.Run("failing to keep an undecodable response is joined to the decoding error", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(blocker, nil, 0o600); err != nil {
			t.Fatal(err)
		}

		srv, _ := recordingServer(t, http.StatusOK, "application/json", cronWithExtra)

		c, err := NewClient(WithBaseURL(srv), WithDebugFile(filepath.Join(blocker, "interactions.json")))
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.Cron(t.Context(), CronParams{})
		if _, ok := errors.AsType[*api.DecodingError](err); !ok {
			t.Fatalf("want *api.DecodingError, got %T: %v", err, err)
		}

		if !strings.Contains(err.Error(), "interactions file") {
			t.Errorf("want the debug file error joined, got: %v", err)
		}
	})

	t.Run("transport error", func(t *testing.T) {
		c, err := NewClient(WithDebugFile(filepath.Join(t.TempDir(), "interactions.json")),
			WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, io.ErrUnexpectedEOF
			})}))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.Cron(t.Context(), CronParams{}); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("want %v, got: %v", io.ErrUnexpectedEOF, err)
		}
	})

	t.Run("WithDebug writes to api/interactions.json", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		srv, _ := recordingServer(t, http.StatusTeapot, "application/json", "")

		c, err := NewClient(WithBaseURL(srv), WithDebug)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.Cron(t.Context(), CronParams{}); !errors.Is(err, api.ErrUnknownStatusCode) {
			t.Fatalf("want %v, got: %v", api.ErrUnknownStatusCode, err)
		}

		if n := len(readInteractions(t, filepath.Join(dir, "api", "interactions.json"))); n != 1 {
			t.Fatalf("want 1 interaction kept, got %d", n)
		}
	})
}

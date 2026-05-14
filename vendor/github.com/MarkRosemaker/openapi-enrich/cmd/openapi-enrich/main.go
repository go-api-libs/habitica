package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/MarkRosemaker/openapi"
	enrich "github.com/MarkRosemaker/openapi-enrich"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
	"github.com/MarkRosemaker/openapi-enrich/recorder"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-enrich: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var specPath, iaPath, auth string
	flag.StringVar(&specPath, "spec", "api/openapi.json", "path to OpenAPI spec file")
	flag.StringVar(&iaPath, "ia", "api/interactions.json", "path to interactions file")
	flag.StringVar(&auth, "auth", "", "authorization header")
	flag.Parse()

	doc, err := openapi.LoadFromFile(specPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		doc = enrich.NewDocument()
	}

	prevIas, err := cassette.InteractionsReadFile(iaPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	tr := recorder.NewTransport(http.DefaultTransport, prevIas)

	// Call requests that don't have a response yet
	scaffoldNext := false
	for _, ia := range prevIas {
		if ia.Response.StatusCode > 0 {
			continue
		}

		if ia.Request.URL == "" {
			scaffoldNext = true
			continue
		}

		req, err := ia.Request.Create(ctx)
		if err != nil {
			return err
		}

		if auth != "" {
			req.Header.Set("Authorization", auth)
		}

		if _, err := tr.RoundTrip(req); err != nil {
			return err
		}
	}

	if err := enrich.Enrich(doc, tr.Interactions); err != nil {
		return err
	}

	// Sort responses and components (but not paths to keep the order)
	for _, path := range doc.Paths {
		for _, op := range path.Operations {
			op.Responses.Sort()
		}
	}
	doc.Components.SortMaps()

	if err := doc.Validate(); err != nil {
		return fmt.Errorf("produced invalid doc: %w", err)
	}

	if err := doc.WriteToFile(specPath); err != nil {
		return err
	}

	ias := tr.Interactions
	ias.Mask()
	ias.TrimResponseHeaders()

	if scaffoldNext {
		ias = append(ias, cassette.Interaction{})
	}

	if err := ias.WriteFile(iaPath); err != nil {
		return err
	}

	return nil
}

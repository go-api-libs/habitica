// Package render generates Go source files from an IR document using embedded templates.
package render

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/imports"
	gofumpt "mvdan.cc/gofumpt/format"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// templateSub is the templates/ subdirectory of the embedded FS, extracted at init.
var templateSub fs.FS

func init() {
	var err error
	templateSub, err = fs.Sub(templateFS, "templates")
	if err != nil {
		panic("openapi-codegen: templates directory missing from embedded FS: " + err.Error())
	}
}

// File represents a single rendered output file.
type File struct {
	Name    string
	Content []byte
}

// Files renders all embedded templates against the IR document.
func Files(doc *ir.Document) ([]File, error) {
	return FilesFromFS(templateSub, doc)
}

// FilesFromFS renders all *.tmpl files found in fsys against the IR document.
// The output file name is the template name minus the ".tmpl" suffix.
func FilesFromFS(fsys fs.FS, doc *ir.Document) ([]File, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	var files []File
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".tmpl" {
			continue
		}

		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}

		rendered, err := RenderTemplate(e.Name(), string(data), doc)
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", e.Name(), err)
		}

		outName := e.Name()[:len(e.Name())-len(".tmpl")]
		files = append(files, File{Name: outName, Content: rendered})
	}

	return files, nil
}

// RenderTemplate executes a template string against the IR document and
// post-processes the result with goimports and gofumpt.
func RenderTemplate(tmplName, tmplContent string, doc *ir.Document) ([]byte, error) {
	tmpl, err := template.New(tmplName).Funcs(templateFuncs()).Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, doc); err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	// goimports: remove unused imports, add missing ones.
	processed, err := imports.Process(tmplName, buf.Bytes(), nil)
	if err != nil {
		return buf.Bytes(), fmt.Errorf("goimports: %w\n---\n%s", err, buf.String())
	}

	// gofumpt: stricter formatting on top of goimports output.
	return gofumpt.Source(processed, gofumpt.Options{
		LangVersion: "go1.25",
		ModulePath:  "github.com/MarkRosemaker/openapi-codegen",
	})
}

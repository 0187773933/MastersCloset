package server

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

// All front-end assets are compiled into the binary. This is the structural fix
// for v1's "ton of stuff is hardcoded": there are no "./v1/server/html/..."
// path literals and no cp -r build step — the binary runs from any directory.
//
//	static/    css / js / images (served verbatim under /static)
//	html/      JS-driven pages served as-is (settings panel, forms)
//	templates/ Go templates rendered with server-side data
//
//go:embed static
var staticFS embed.FS

//go:embed html
var htmlFS embed.FS

//go:embed templates
var templatesFS embed.FS

// templates holds the parsed dynamic templates.
var templates = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

// staticHTTPFS is the /static document root.
func staticHTTPFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// page returns the bytes of a raw html/ page.
func page(name string) ([]byte, error) {
	return htmlFS.ReadFile("html/" + name)
}

/*
Package wutrender helps to render templates.

Example:
	// main.go
	package main

	import (
	  "github.com/8protons/wutrender"
	)

	func main() {
	  // Init default Renderer with default options
	  wutrender.Init()

	  // Init default Renderer with custom options
	  wutrender.Init(wutrender.Options{
	    Directory: "app/templates",
	    Layout:    "layout",
	  })

	  // return (*bytes.Buffer, error)
	  wutrender.HTML("sessions/new", map[string]interface{}{ "hello": "world" })
	}

	// ..
	// WriteHTML is an easy way to write to ResponseWriter
	func IndexHandler(w http.ResponseWriter, r *http.Request) {
	  data := map[string]interface{}{}
	  data["counter"] = 5

	  wutrender.WriteHTML(w, 200, "hello", data)
	}
	// ..

By default the `wutrender.Renderer` will load templates with *.tmpl extension from the "templates" directory.

It uses `filepath/filename.{format}.{extension}` scheme to distinguish folders and content formats.

So file with "templates/sessions/new.html.tmpl" path will be available to render with "sessions/new.html" name.

Layouts and partials support.

wutrender requires Go 1.2 or newer.
*/
package wutrender

import (
	"bytes"
	"fmt"
	"github.com/8protons/wutenv"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	ContentType = "Content-Type"
	ContentJSON = "application/json; charset=utf-8"
	ContentHTML = "text/html; charset=utf-8"
	ContentJS   = "application/javascript; charset=utf-8"
)

// Helper functions placeholders
var helperFunctions = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called without layout")
	},
	"partial": func(name string, binding ...interface{}) (string, error) {
		return "", fmt.Errorf("partial called without implementation")
	},
}

// Delims represents a set of Left and Right delimiters for HTML template rendering
type Delims struct {
	// Left delimiter, defaults to {{
	Left string
	// Right delimiter, defaults to }}
	Right string
}

type Options struct {
	// Directory to load templates. Default is "templates"
	Directory string
	// Layout template name. Will not render a layout if "". Defaults to "".
	Layout string
	// Extensions to parse template files from. Defaults to [".tmpl"]
	Extensions []string
	// Template delimiters
	Delims Delims
	// Helper functions. Defaults to [].
	Funcs []template.FuncMap
}

// Renderer struct
type Renderer struct {
	t       *template.Template
	options Options
}

// Template copy - has all rendering methods
type TemplateCopy struct {
	t      *template.Template
	layout string
}

func New(opt ...Options) *Renderer {
	options := prepareOptions(opt)
	r := &Renderer{
		options: options,
	}

	r.t = r.compile()

	return r
}

// Default Renderer options
func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

	// Defaults
	if len(opt.Directory) == 0 {
		opt.Directory = "templates"
	}
	if len(opt.Extensions) == 0 {
		opt.Extensions = []string{".tmpl"}
	}

	return opt
}

func (r *Renderer) compile() *template.Template {
	t := template.New(r.options.Directory)

	t.Delims(r.options.Delims.Left, r.options.Delims.Right)

	template.Must(t.Parse("wut!"))

	filepath.Walk(r.options.Directory, func(path string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(r.options.Directory, path)
		if err != nil {
			return err
		}

		fileExt := filepath.Ext(relPath)

		for _, v := range r.options.Extensions {
			if v == fileExt {

				// Read file and panic on error
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := strings.TrimSuffix(relPath, filepath.Ext(relPath))
				tmpl := t.New(filepath.ToSlash(name))

				// add our funcmaps
				for _, funcs := range r.options.Funcs {
					t.Funcs(funcs)
				}

				t.Funcs(helperFunctions)

				// template.Must(tmpl.Parse(string(buf)))
				_, err = tmpl.Parse(string(buf))
				if err != nil {
					panic(err)
				}
				break
			}
		}

		return nil
	}) // end Walk

	return t
}

// Return *TemplateCopy to guarantee cleanness of the source templates.
func (r *Renderer) Copy() *TemplateCopy {
	var tc *template.Template

	// Recompile template
	if wutenv.IsDev {
		tc = r.compile()
	} else {
		var err error
		tc, err = r.t.Clone()

		if err != nil {
			panic(err)
		}
	}

	return &TemplateCopy{
		t:      tc,
		layout: r.options.Layout,
	}
}

// Render HTML with layout support
func (tmpl *TemplateCopy) HTML(name string, binding interface{}) (*bytes.Buffer, error) {
	return tmpl.RenderFormat("html", name, binding)
}

// Write HTML to ResponseWriter
func (tmpl *TemplateCopy) WriteHTML(rw http.ResponseWriter, status int, name string, binding interface{}) {
	html, err := tmpl.HTML(name, binding)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set(ContentType, ContentHTML)
	rw.WriteHeader(status)
	rw.Write(html.Bytes())
}

// Shortcut for RenderFormat("js", ...) - render Javascript file
func (tmpl *TemplateCopy) JS(name string, binding interface{}) (*bytes.Buffer, error) {
	return tmpl.RenderFormat("js", name, binding)
}

// Write JS file to ResponseWriter
func (tmpl *TemplateCopy) WriteJS(rw http.ResponseWriter, status int, name string, binding interface{}) {
	html, err := tmpl.RenderFormat("js", name, binding)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set(ContentType, ContentJS)
	rw.WriteHeader(status)
	rw.Write(html.Bytes())
}

// General function to render template with "name.{format}" scheme
func (tmpl *TemplateCopy) RenderFormat(format string, name string, binding interface{}) (*bytes.Buffer, error) {

	// Add partial support
	addPartial(tmpl.t)

	fullName := name + "." + format

	// Set yield function (layout)
	if format == "html" && tmpl.layout != "" {
		addYield(tmpl.t, fullName, binding)
		fullName = tmpl.layout + ".html"
	}

	return executeTemplate(tmpl.t, fullName, binding)
}

// Override default layout
func (tmpl *TemplateCopy) SetLayout(layout string) *TemplateCopy {
	tmpl.layout = layout

	return tmpl
}

// Set template.FuncMap - it's safe and does not change source templates
func (tmpl *TemplateCopy) SetFuncs(funcs template.FuncMap) *TemplateCopy {
	tmpl.t.Funcs(funcs)

	return tmpl
}

// Add yield keyword
func addYield(t *template.Template, name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := executeTemplate(t, name, binding)
			// return safe html here since we are rendering our own template
			return template.HTML(buf.String()), err
		},
	}
	t.Funcs(funcs)
}

// Add partial keyword
func addPartial(t *template.Template) {
	funcs := template.FuncMap{
		"partial": func(name string, pairs ...interface{}) (template.HTML, error) {
			binding, err := mapFromPairs(pairs...)

			if err != nil {
				return "", err
			}

			dir, filename := filepath.Split(name)

			buf, err := executeTemplate(t, dir+"_"+filename+".html", binding)

			// return safe html
			return template.HTML(buf.String()), err
		},
	}
	t.Funcs(funcs)
}

// mapFromPairs converts interface parameters to a string map for partial binding
func mapFromPairs(pairs ...interface{}) (interface{}, error) {
	length := len(pairs)

	if length == 1 {
		return pairs[0], nil
	}

	if length%2 != 0 {
		return nil, fmt.Errorf("wutrender: number of parameters must be multiple of 2, got %v", pairs)
	}

	m := make(map[string]interface{}, length/2)

	for i := 0; i < length; i += 2 {
		v := pairs[i]
		switch v.(type) {
		case string:
			m[v.(string)] = pairs[i+1]
		default:
			return nil, fmt.Errorf("wutrender: pairs should be in format \"string => interface{}\", got %v, %v", v, pairs[i+1])
		}
	}
	return m, nil
}

func executeTemplate(t *template.Template, name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := t.ExecuteTemplate(buf, name, binding)

	if err != nil {
		return bytes.NewBufferString(err.Error()), err
	}

	return buf, nil
}

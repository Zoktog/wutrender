package wutrender

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewRenderer(t *testing.T) {
	r := New(Options{
		Directory: "fixtures",
	})

	assert.Equal(t, len(r.t.Templates()), 3)
}

func Test_TemplateNames(t *testing.T) {
	r := New(Options{
		Directory: "fixtures",
	})

	assert.NotEqual(t, r.t.Lookup("base/hello"), nil)
	assert.Nil(t, r.t.Lookup("base/notemplate"))
}

func Test_HTML(t *testing.T) {
	r := New(Options{
		Directory: "fixtures",
	})

	html, _ := r.Copy().HTML("base/hello", nil)
	htmlBind, _ := r.Copy().HTML("base/hello", []string{"willkommen"})

	assert.Equal(t, html.String(), "<div>Hello </div>")
	assert.Equal(t, htmlBind.String(), "<div>Hello [willkommen]</div>")
}

func Test_LayoutHTML(t *testing.T) {
	r := New(Options{
		Directory: "fixtures",
		Layout:    "base/layout",
	})

	html, err := r.Copy().HTML("base/hello", nil)
	assert.Nil(t, err)
	htmlBind, _ := r.Copy().HTML("base/hello", []string{"willkommen"})

	assert.Equal(t, html.String(), "head\n<div>Hello </div>\nfoot")
	assert.Equal(t, htmlBind.String(), "head\n<div>Hello [willkommen]</div>\nfoot")
}

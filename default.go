package wutrender

import (
	"bytes"
	"net/http"
)

var DefaultRenderer *Renderer

func Init(opts ...Options) {
	DefaultRenderer = New(opts...)
}

func Copy() *TemplateCopy {
	if DefaultRenderer == nil {
		panic("You should call wutrender.Init(opts ...Options) first")
	}

	return DefaultRenderer.Copy()
}

func HTML(name string, binding interface{}) (*bytes.Buffer, error) {
	if DefaultRenderer == nil {
		panic("You should call wutrender.Init(opts ...Options) first")
	}

	return DefaultRenderer.Copy().HTML(name, binding)
}

func WriteHTML(rw http.ResponseWriter, status int, name string, binding interface{}) {
	if DefaultRenderer == nil {
		panic("You should call wutrender.Init(opts ...Options) first")
	}

	DefaultRenderer.Copy().WriteHTML(rw, status, name, binding)
}

func JS(name string, binding interface{}) (*bytes.Buffer, error) {
	if DefaultRenderer == nil {
		panic("You should call wutrender.Init(opts ...Options) first")
	}

	return DefaultRenderer.Copy().RenderFormat("js", name, binding)
}

func WriteJS(rw http.ResponseWriter, status int, name string, binding interface{}) {
	if DefaultRenderer == nil {
		panic("You should call wutrender.Init(opts ...Options) first")
	}

	DefaultRenderer.Copy().WriteJS(rw, status, name, binding)
}

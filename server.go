package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"text/template"
	"time"
)

//go:embed server/templates/*
var templateFS embed.FS

//go:embed server/static/*
var staticFS embed.FS

type server struct {
	indexTemplate   *template.Template
	listener        string
	defaultRenderer string
	configfile      string
	base            string
	glob            string
	cache           cache
}

func NewServer(listener, defaultRenderer, configfile, base, glob, cacheTimeout string) (*server, error) {
	durr, err := time.ParseDuration(cacheTimeout)
	if err != nil {
		return nil, err
	}
	s := &server{
		listener:        listener,
		defaultRenderer: defaultRenderer,
		configfile:      configfile,
		base:            base,
		glob:            glob,
		cache:           *NewCache(durr),
	}
	s.indexTemplate, err = template.ParseFS(templateFS, "server/templates/index.html.tmpl")
	return s, err
}

func (s *server) Run() error {
	http.HandleFunc("/index.html", s.HandleIndex)

	assets, err := fs.Sub(staticFS, "server")
	if err != nil {
		return err
	}
	fs := http.FileServer(http.FS(assets))
	http.Handle("/static/", fs)

	http.HandleFunc("/", s.HandleIndex)

	fmt.Printf("server listening on http://%s/, hit CTRL-C to stop server...\n", s.listener)
	err = http.ListenAndServe(s.listener, nil)
	return err
}

func (s *server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	key := r.URL.RawQuery
	if cached, ok := s.cache.Get(key); ok {
		_, _ = w.Write(cached)
		return
	}
	focusElems := r.URL.Query().Get("focus")
	focus := strings.Split(focusElems, "_")

	rendererName := r.URL.Query().Get("renderer")
	if rendererName == "" {
		rendererName = s.defaultRenderer
	}

	cfg, err := NewConfig(s.configfile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// build system
	sys, errs := build(s.base, s.glob, focus)
	if len(errs) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		out := ""
		for _, err = range errs {
			out += fmt.Sprintf("%s\n", err.Error())
		}
		_, _ = w.Write([]byte(out))
		return
	}

	// render template
	renderer, ok := cfg.Renderer[rendererName]
	if !ok {
		exitOnErr(fmt.Errorf("renderer %s not specified in %s", rendererName, s.configfile))
	}
	out, err := render(sys, renderer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// create svg
	img, err := svg(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	var buf bytes.Buffer

	thing := strings.TrimSuffix(strings.TrimPrefix(string(img), `<?xml version="1.0" encoding="utf-8"?><svg`), "</xml>")
	thing = `<svg id="svg" class="svg"` + thing
	err = s.indexTemplate.Execute(&buf, thing)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	data := buf.Bytes()
	s.cache.Add(key, data)
	_, _ = w.Write(data)
}

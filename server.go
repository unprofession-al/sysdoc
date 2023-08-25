package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sysdoc/internal/cache"
	"sysdoc/internal/persistence"
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
	cache           cache.Cache
	persistence     persistence.Persistence
	renderer        Renderer
}

func NewServer(listener, defaultRenderer, configfile, base, glob, cacheTimeout string, p persistence.Persistence, r Renderer) (*server, error) {
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
		cache:           *cache.New(durr),
		persistence:     p,
		renderer:        r,
	}
	s.indexTemplate, err = template.ParseFS(templateFS, "server/templates/index.html.tmpl")
	return s, err
}

func (s *server) Run() error {
	assets, err := fs.Sub(staticFS, "server")
	if err != nil {
		return err
	}
	fs := http.FileServer(http.FS(assets))
	http.Handle("/static/", fs)
	http.HandleFunc("/svg/", s.HandleSVG)
	http.HandleFunc("/branches.json", s.HandleBranches)
	http.HandleFunc("/index.html", s.HandleIndex)
	http.HandleFunc("/", s.HandleIndex)

	fmt.Printf("server listening on http://%s/, hit CTRL-C to stop server...\n", s.listener)
	err = http.ListenAndServe(s.listener, nil)
	return err
}

func (s *server) HandleBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := s.persistence.Branches()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(branches)
}

func (s *server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	err := s.indexTemplate.Execute(&buf, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	data := buf.Bytes()
	_, _ = w.Write(data)
}

func (s *server) HandleSVG(w http.ResponseWriter, r *http.Request) {
	key := r.URL.RawQuery
	if cached, ok := s.cache.Get(key); ok {
		_, _ = w.Write(cached)
		return
	}
	focusElems := r.URL.Query().Get("focus")
	focus := strings.Split(focusElems, "_")

	branch := r.URL.Query().Get("branch")
	if branch != "" {
		err := s.persistence.CheckoutBranch(branch)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}

	rendererName := r.URL.Query().Get("renderer")
	if rendererName == "" {
		rendererName = s.defaultRenderer
	}

	cfg, err := NewConfig(s.configfile, s.persistence.Filesystem())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	// build system
	sys, errs := New(s.base, s.glob, focus, s.persistence)
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
	img, err := s.renderer.Do(sys, renderer, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tidy := strings.TrimSuffix(strings.TrimPrefix(string(img), `<?xml version="1.0" encoding="utf-8"?><svg`), "</xml>")
	tidy = `<svg id="svg" class="svg"` + tidy

	s.cache.Add(key, []byte(tidy))
	w.Header().Set("content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(tidy))
}

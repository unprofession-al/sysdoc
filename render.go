package main

import (
	"bytes"
	"fmt"
	"sysdoc/internal/postprocessor"
	"sysdoc/internal/postprocessor/d2"
	"text/template"
)

// ELEMENT

type ElementTemplateData struct {
	data  *element
	templ string
}

func newElementTemplateData(e *element, t string) ElementTemplateData {
	return ElementTemplateData{data: e, templ: t}
}
func (e ElementTemplateData) ID(sep string) string    { return e.data.getID(sep) }
func (e ElementTemplateData) Fragment() string        { return e.data.fragment }
func (e ElementTemplateData) Name() string            { return e.data.name }
func (e ElementTemplateData) Tags() map[string]string { return e.data.tags }
func (e ElementTemplateData) Doc() string             { return string(e.data.doc) }
func (e ElementTemplateData) Children() []string {
	out := []string{}
	for _, child := range e.data.children {
		er, _ := newElementTemplateData(child, e.templ).render()
		out = append(out, er)
	}
	return out
}
func (e ElementTemplateData) Interfaces() []InterfaceTemplateData {
	out := []InterfaceTemplateData{}
	for _, i := range e.data.interfaces {
		out = append(out, newInterfaceTemplateData(i, ""))
	}
	return out
}
func (e ElementTemplateData) Propagations() []InterfaceTemplateData {
	out := []InterfaceTemplateData{}
	for _, i := range e.data.propagations {
		out = append(out, newInterfaceTemplateData(i, ""))
	}
	return out
}
func (e ElementTemplateData) render() (string, error) {
	var b bytes.Buffer
	t, err := template.New("tmpl").Parse(e.templ)
	if err != nil {
		return "", err
	}
	err = t.Execute(&b, e)
	return b.String(), err
}

// INTERFACE

type InterfaceTemplateData struct {
	data  *interf
	templ string
}

func newInterfaceTemplateData(i *interf, t string) InterfaceTemplateData {
	return InterfaceTemplateData{data: i, templ: t}
}
func (i InterfaceTemplateData) Name() string                   { return i.data.name }
func (i InterfaceTemplateData) Fragment() string               { return i.data.fragment }
func (i InterfaceTemplateData) ID(sep string) string           { return i.data.getID(sep) }
func (i InterfaceTemplateData) PropagatesID(sep string) string { return i.data.propagates.getID(sep) }
func (i InterfaceTemplateData) Tags() map[string]string        { return i.data.tags }
func (i InterfaceTemplateData) render() (string, error) {
	var b bytes.Buffer
	t, err := template.New("tmpl").Parse(i.templ)
	if err != nil {
		return "", err
	}
	err = t.Execute(&b, i)
	return b.String(), err
}

// DEPENDENCY

type DependencyTemplateData struct {
	data  *dependency
	templ string
}

func newDependencyTemplateData(d *dependency, t string) DependencyTemplateData {
	return DependencyTemplateData{data: d, templ: t}
}
func (d DependencyTemplateData) render() (string, error) {
	var b bytes.Buffer
	t, err := template.New("tmpl").Parse(d.templ)
	if err != nil {
		return "", err
	}
	err = t.Execute(&b, d)
	return b.String(), err
}
func (d DependencyTemplateData) Fragment() string              { return d.data.fragment }
func (d DependencyTemplateData) Description() string           { return d.data.description }
func (d DependencyTemplateData) Tags() map[string]string       { return d.data.tags }
func (d DependencyTemplateData) BelongsToID(sep string) string { return d.data.belongsTo.getID(sep) }
func (d DependencyTemplateData) DependsOnID(sep string) string { return d.data.dependsOn.getID(sep) }
func (d DependencyTemplateData) ViaPropagation(sep string) string {
	return d.data.viaPropagation.getID(sep)
}

// RENDER

type Renderer struct {
	postprocessors map[string]func(postprocessor.Config) (postprocessor.Postprocessor, error)
}

func NewRenderer() *Renderer {
	r := &Renderer{postprocessors: map[string]func(postprocessor.Config) (postprocessor.Postprocessor, error){}}
	r.postprocessors["d2"] = d2.New
	return r
}

func (r *Renderer) Do(e *element, rc renderConfig, noPostprocessor bool) ([]byte, error) {
	data := struct {
		Elements     string
		Dependencies string
		Propagations string
	}{}

	er, err := newElementTemplateData(e, rc.Templates.Element).render()
	if err != nil {
		return nil, err
	}
	data.Elements += er

	for _, dep := range e.getDependencies() {
		dr, err := newDependencyTemplateData(dep, rc.Templates.Dependency).render()
		if err != nil {
			return nil, err
		}
		data.Dependencies += dr
	}

	for _, prop := range e.getPropagations() {
		pr, err := newInterfaceTemplateData(prop, rc.Templates.Propagation).render()
		if err != nil {
			return nil, err
		}
		data.Propagations += pr
	}

	var b bytes.Buffer
	t, err := template.New("tmpl").Parse(rc.Templates.Global)
	if err != nil {
		return nil, err
	}

	err = t.Execute(&b, data)
	if noPostprocessor || rc.Postprocessor.Name == "" {
		return b.Bytes(), err
	}

	initPostprocessor, ok := r.postprocessors[rc.Postprocessor.Name]
	if !ok {
		list := []string{}
		for key := range r.postprocessors {
			list = append(list, key)
		}
		return nil, fmt.Errorf("No postprocessor with name '%s' available, please choose one of %v...", rc.Postprocessor.Name, list)
	}
	postprocessor, err := initPostprocessor(rc.Postprocessor)
	if err != nil {
		return nil, err
	}

	return postprocessor.Process(b.String())
}

type renderConfig struct {
	Templates struct {
		Element     string `yaml:"element"`
		Dependency  string `yaml:"dependency"`
		Propagation string `yaml:"propagation"`
		Global      string `yaml:"global"`
	} `yaml:"templates"`
	Postprocessor postprocessor.Config
}

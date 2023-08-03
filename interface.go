package main

import (
	"fmt"
	"strings"
)

type interfConfiguration struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Tags        map[string]string `yaml:"tags" json:"tags"`
}

type interf struct {
	fragment    string
	name        string
	description string
	tags        map[string]string

	belongsTo  *element
	propagates *interf
	k          bool
}

func newInterf(fragment string, c interfConfiguration, e *element) *interf {
	return &interf{
		fragment:    fragment,
		name:        c.Name,
		description: c.Description,
		tags:        c.Tags,
		belongsTo:   e,
	}
}

func (i *interf) propagateTo(e *element) *interf {
	if i.belongsTo == e {
		return i
	}
	parent := i.belongsTo.parent
	if parent == nil {
		return nil
	}
	prop := &interf{
		fragment:    fmt.Sprintf("%s-%s", i.fragment, i.belongsTo.fragment),
		name:        fmt.Sprintf("%s (propagated from %s)", i.name, i.belongsTo.name),
		description: i.description,
		belongsTo:   parent,
		propagates:  i,
	}
	exists := false
	for _, p := range parent.propagations {
		if p.fragment == prop.fragment {
			prop = p
			exists = true
		}
	}
	if !exists {
		parent.propagations = append(parent.propagations, prop)
	}
	if e == parent {
		return prop
	}
	return prop.propagateTo(e)
}

func (i *interf) getID(join string) string {
	return strings.Join(append([]string{i.belongsTo.getID(join)}, i.fragment), join)
}

func (i *interf) keep() {
	i.k = true
	i.belongsTo.keep()
}

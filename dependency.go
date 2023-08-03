package main

import (
	"strings"
)

type dependencyConfiguration struct {
	DependsOn   string            `yaml:"depends_on" json:"depends_on"`
	Description string            `yaml:"description" json:"description"`
	Tags        map[string]string `yaml:"tags" json:"tags"`
}

type dependency struct {
	fragment    string
	description string
	reference   string
	tags        map[string]string

	dependsOn      *interf
	viaPropagation *interf
	belongsTo      *element
	k              bool
}

func newDependency(fragment string, c dependencyConfiguration, e *element) *dependency {
	return &dependency{
		fragment:    fragment,
		description: c.Description,
		reference:   c.DependsOn,
		tags:        c.Tags,
		belongsTo:   e,
	}
}

func (d *dependency) positionFromReference() []string {
	return strings.Split(d.reference, ".")
}

func (d *dependency) keep() {
	d.k = true
	d.dependsOn.keep()
	d.belongsTo.keep()
}

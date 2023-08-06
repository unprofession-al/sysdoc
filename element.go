package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/frontmatter"
)

type elementConfiguration struct {
	Name         string                             `yaml:"name" json:"name"`
	Tags         map[string]string                  `yaml:"tags" json:"tags"`
	Dependencies map[string]dependencyConfiguration `yaml:"dependencies" json:"dependencies"`
	Interfaces   map[string]interfConfiguration     `yaml:"interfaces" json:"interfaces"`
	Doc          []byte                             `yaml:"-" json:"-"`
}

func newElementConfigurationFromFile(path string) (elementConfiguration, error) {
	ec := elementConfiguration{}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ec, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Could not read file '%s', error occured: %w", path, err)
		return ec, err
	}
	reader := bytes.NewReader(data)
	doc, err := frontmatter.Parse(reader, &ec)
	if err != nil {
		err = fmt.Errorf("Could not parse data of '%s', error occured: %w", path, err)
		return ec, err
	}
	ec.Doc = doc
	return ec, err
}

type element struct {
	// configured values
	fragment string
	name     string
	tags     map[string]string
	doc      []byte

	// calculated values
	dependencies []*dependency
	interfaces   []*interf
	propagations []*interf
	children     []*element
	parent       *element
	k            bool
}

func newElement(fragment string, c elementConfiguration) *element {
	e := &element{
		fragment: fragment,
		name:     c.Name,
		tags:     c.Tags,
		doc:      c.Doc,
	}
	for key, dep := range c.Dependencies {
		e.dependencies = append(e.dependencies, newDependency(key, dep, e))
	}
	for key, interf := range c.Interfaces {
		e.interfaces = append(e.interfaces, newInterf(key, interf, e))
	}
	return e
}

func positionFromID(id, devider string) []string {
	seg := strings.Split(id, devider)
	out := []string{}
	for _, s := range seg {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getPosition(basepath, path string) []string {
	basepath = filepath.Clean(basepath)
	path = filepath.Clean(path)
	//basepath = filepath.Dir(basepath)
	path = strings.TrimPrefix(path, basepath)
	path = strings.Trim(path, string(os.PathSeparator))
	pos := strings.Split(path, string(os.PathSeparator))
	if len(pos) == 1 && pos[0] == "" {
		return []string{}
	}
	return pos
}

func newElementFromFS(basepath, matcher string) (*element, error) {
	basepath = filepath.Clean(basepath)
	_, err := os.Stat(basepath)
	if err != nil {
		return nil, err
	}

	// read all configuration files
	configs := map[string]elementConfiguration{}
	err = filepath.WalkDir(basepath, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			err = fmt.Errorf("Could not stat file '%s', error occured: %w", path, err)
			return err
		}
		if !info.IsDir() {
			return nil
		}
		c, err := newElementConfigurationFromFile(filepath.Join(path, matcher))
		if err != nil {
			return err
		}
		configs[filepath.Clean(path)] = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	// generate element tree from configurations
	e := &element{}
	inited := false
	// sort by path
	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// create elements and add to tree
	for _, k := range keys {
		pos := getPosition(basepath, k)
		fragment := "root"
		if len(pos) > 0 {
			fragment = pos[len(pos)-1]
		}
		elem := newElement(fragment, configs[k])
		if !inited {
			e = elem
			inited = true
		} else {
			err = e.appendAt(elem, pos)
			if err != nil {
				return e, err
			}
		}
	}

	return e, nil
}

// appendAt adds a child element on the root element, where the path indicates each fragment of
// the parent elements and the element itself
func (e *element) appendAt(a *element, pos []string) error {
	if len(pos) == 1 {
		e.children = append(e.children, a)
		a.parent = e
		return nil
	} else if len(pos) > 1 {
		for _, child := range e.children {
			if child.fragment == pos[0] {
				trimmed := pos[1:]
				return child.appendAt(a, trimmed)
			}
		}
	}
	return fmt.Errorf("Could not append '%s' to '%s'", strings.Join(pos, "/"), strings.Join(e.position(), "/"))
}

func (e *element) position() []string {
	out := []string{e.fragment}
	if e.parent != nil {
		out = append(e.parent.position(), out...)
	}
	return out
}

func (e *element) getID(join string) string {
	return strings.Join(e.position()[1:], join)
}

func (e *element) findElementByPosition(pos []string) (*element, error) {
	if len(pos) == 0 {
		return e, nil
	}
	for _, elem := range e.children {
		// skip loop if no match
		if elem.fragment != pos[0] {
			continue
		}
		// return element if position is onsy one element
		if len(pos) == 1 {
			return elem, nil
		}
		// go deeper if more than one element
		next, err := elem.findElementByPosition(pos[1:])
		if err != nil {
			break
		}
		return next, nil

	}
	return nil, fmt.Errorf("Element '%s' not found", strings.Join(pos, "."))
}

func (e *element) findInterfaceByPosition(pos []string) (*interf, error) {
	if len(pos) < 1 {
		return nil, fmt.Errorf("Position too short")
	}
	elemPos := pos[:len(pos)-1]
	elem, err := e.findElementByPosition(elemPos)
	if err != nil {
		return nil, fmt.Errorf("Error while finding interface '%s': %w", strings.Join(pos, "."), err)
	}
	for _, i := range elem.interfaces {
		if i.fragment == pos[len(pos)-1] {
			return i, nil
		}
	}
	return nil, fmt.Errorf("Element '%s' does not provide interface '%s'", strings.Join(elemPos, "."), strings.Join(pos, "."))
}

func (e *element) resolveDependencies(root *element) []error {
	errs := []error{}
	for _, dep := range e.dependencies {
		pos := dep.positionFromReference()
		i, err := root.findInterfaceByPosition(pos)
		if err != nil {
			err = fmt.Errorf("Could not resolve dependency '%s' of element '%s': %w", dep.reference, strings.Join(e.position(), "."), err)
			errs = append(errs, err)
		}
		dep.dependsOn = i
	}
	for _, child := range e.children {
		childErrs := child.resolveDependencies(root)
		errs = append(errs, childErrs...)
	}
	return errs
}

func (e *element) getDependencies() []*dependency {
	out := e.dependencies
	for _, elem := range e.children {
		out = append(out, elem.getDependencies()...)
	}
	return out
}

func (e *element) getPropagations() []*interf {
	out := []*interf{}
	out = append(out, e.propagations...)
	for _, elem := range e.children {
		out = append(out, elem.getPropagations()...)
	}
	return out
}

func (e *element) propagateInterfaces() error {
	for _, dep := range e.dependencies {
		i := dep.dependsOn
		if i == nil {
			return fmt.Errorf("Dependencies of '%s' are not yet resolved", e.getID("."))
		}
		sibling := e.closestSibling(i.belongsTo)
		if sibling == nil {
			return fmt.Errorf("Could not propagate interface '%s' of '%s', no sibling found", i.name, i.belongsTo.getID("."))
		}
		// do not propagate to root
		if sibling.parent != nil {
			prop := i.propagateTo(sibling)
			dep.viaPropagation = prop
		} else {
			dep.viaPropagation = i
		}
	}
	for _, child := range e.children {
		err := child.propagateInterfaces()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *element) hasParent(b *element) bool {
	if len(b.position()) >= len(a.position()) {
		return false
	}
	if a.parent == b {
		return true
	}
	return a.parent.hasParent(b)
}

func (a *element) sharedParent(b *element) *element {
	// determine lenght of shortest position
	shortest, longest := &element{}, &element{}
	if len(a.position()) < len(b.position()) {
		shortest = a
		longest = b
	} else {
		shortest = b
		longest = a
	}
	if len(shortest.position()) == 0 {
		return nil
	}
	if longest.hasParent(shortest) {
		return shortest
	}
	return longest.sharedParent(shortest.parent)
}

func (a *element) closestSibling(b *element) *element {
	if a.parent == b.parent {
		return b
	}
	sp := a.sharedParent(b)
	if sp == b.parent {
		return sp
	}
	for _, e := range sp.children {
		if b.hasParent(e) {
			return e
		}
	}
	return nil
}

func (e *element) keep() {
	e.k = true
	if e.parent != nil {
		e.parent.keep()
	}
}

func (root *element) focus(pos []string) error {
	elems := []*element{}
	for _, p := range pos {
		e, err := root.findElementByPosition(positionFromID(p, "."))
		if err != nil {
			return err
		}
		elems = append(elems, e)
	}
	for _, e := range elems {
		root.setKeep(e)
	}
	root.tidy()
	return nil
}

func (base *element) setKeep(k *element) {
	if base == k {
		if base.tags == nil {
			base.tags = map[string]string{}
		}
		base.tags["focussed"] = "true"
	}
	if base == k || base.hasParent(k) {
		base.keep()
		for _, d := range base.dependencies {
			d.keep()
		}
		for _, i := range base.interfaces {
			i.keep()
		}
	}
	for _, d := range base.dependencies {
		if d.dependsOn.belongsTo == k || d.dependsOn.belongsTo.hasParent(k) {
			d.keep()
		}
	}
	for _, c := range base.children {
		c.setKeep(k)
	}
}

func (e *element) tidy() {
	dependencies := []*dependency{}
	for key := range e.dependencies {
		if e.dependencies[key].k {
			dependencies = append(dependencies, e.dependencies[key])
		}
	}
	e.dependencies = dependencies

	interfaces := []*interf{}
	for key := range e.interfaces {
		if e.interfaces[key].k {
			interfaces = append(interfaces, e.interfaces[key])
		}
	}
	e.interfaces = interfaces

	children := []*element{}
	for key := range e.children {
		if e.children[key].k {
			e.children[key].tidy()
			children = append(children, e.children[key])
		}
	}
	e.children = children
}
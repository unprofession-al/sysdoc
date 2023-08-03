package main

import (
	"context"
	"fmt"
	"os"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

func main() {
	err := NewApp().Execute()
	exitOnErr(err)
}

func build(basedir, glob string, focus []string) (*element, []error) {
	sys, err := newElementFromFS(basedir, glob)
	if err != nil {
		return sys, []error{err}
	}

	errs := sys.resolveDependencies(sys)
	if len(errs) > 0 {
		return sys, errs
	}

	if len(focus) > 0 {
		err = sys.focus(focus)
		if err != nil {
			return sys, []error{err}
		}
	}

	err = sys.propagateInterfaces()
	if err != nil {
		return sys, []error{err}
	}
	return sys, nil
}

func svg(code string) ([]byte, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return []byte{}, err
	}
	defaultLayout := func(ctx context.Context, g *d2graph.Graph) error {
		return d2elklayout.Layout(ctx, g, nil)
	}
	diagram, _, err := d2lib.Compile(context.Background(), code, &d2lib.CompileOptions{
		Layout: defaultLayout,
		Ruler:  ruler,
	})
	if err != nil {
		return []byte{}, err
	}
	return d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad:     d2svg.DEFAULT_PADDING,
		ThemeID: d2themescatalog.GrapeSoda.ID,
	})
}

func exitOnErr(errs ...error) {
	errNotNil := false
	for _, err := range errs {
		if err == nil {
			continue
		}
		errNotNil = true
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	}
	if errNotNil {
		fmt.Print("\n")
		os.Exit(-1)
	}
}

package d2

import (
	"context"
	"os"
	"sysdoc/internal/postprocessor"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type Postprocessor struct{}

func New(config postprocessor.Config) (postprocessor.Postprocessor, error) {
	return &Postprocessor{}, nil
}

func (p *Postprocessor) Process(code string) ([]byte, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return []byte{}, err
	}
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}
	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}
	ctx := log.With(context.Background(), slog.Make(sloghuman.Sink(os.Stdout)))
	diagram, _, err := d2lib.Compile(ctx, code, compileOpts, renderOpts)
	if err != nil {
		return []byte{}, err
	}
	return d2svg.Render(diagram, renderOpts)
}

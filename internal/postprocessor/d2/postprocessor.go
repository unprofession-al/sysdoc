package d2

import (
	"context"
	"sysdoc/internal/postprocessor"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
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

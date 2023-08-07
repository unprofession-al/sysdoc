package main

import (
	"fmt"
	"os"

	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

type App struct {
	flags struct {
		configfile string
		base       string
		glob       string
		focus      []string
		render     struct {
			renderer string
		}
		svg struct {
			renderer string
			out      string
		}
		serve struct {
			renderer string
			listener string
		}
	}

	// entry point
	Execute func() error
}

func NewApp() *App {
	a := &App{}
	appName := "sysdoc"

	// root
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: "sysdoc allows to document dependencies between systems",
	}
	rootCmd.PersistentFlags().StringVar(&a.flags.configfile, "config", "./sysdoc.yaml", "configuration file path")
	rootCmd.PersistentFlags().StringVar(&a.flags.base, "base", ".", "base directory of the sysdoc definitions")
	rootCmd.PersistentFlags().StringVar(&a.flags.glob, "glob", "README.md", "glob to find sysdoc definitions")
	rootCmd.PersistentFlags().StringSliceVar(&a.flags.focus, "focus", []string{}, "elements to be focussed")
	a.Execute = rootCmd.Execute

	// render
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "renders system documentation in a given template to standard output",
		Run:   a.renderCmd,
	}
	renderCmd.PersistentFlags().StringVar(&a.flags.render.renderer, "renderer", "default", "name of the renederer (set of templates in the configuration file)")
	rootCmd.AddCommand(renderCmd)

	// svg
	svgCmd := &cobra.Command{
		Use:   "svg",
		Short: "renders system documentation in a given d2lang template to a svg file",
		Run:   a.svgCmd,
	}
	svgCmd.PersistentFlags().StringVar(&a.flags.svg.renderer, "renderer", "default", "name of the renederer (set of templates in the configuration file)")
	svgCmd.PersistentFlags().StringVar(&a.flags.svg.out, "out", "sysdoc.svg", "name of the file to be written")
	rootCmd.AddCommand(svgCmd)

	// serve
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "renders system documentation in a given d2lang template to a svg file and serves it over http",
		Long: `With the subcommnad 'serve', a small web server is launched locally. Access the server via
browser and get a rendered SVG picture of your system architecture. The server provides 
a simple API to change the focus as well as the renderer on the fly:

- To change the focus provide a list of elements, where every element is separated with a '+'.
- To switch the renderer provide the 'renderer' query parameter.

For example, focus on the elements 'A.AB' and 'C' and render the output with a renderer
called 'custom' using the following URL: http://localhost:8080/A.AB+C?renderer=custom`,
		Run: a.serveCmd,
	}
	serveCmd.PersistentFlags().StringVar(&a.flags.serve.listener, "listener", "127.0.0.1:8080", "listener to be used by the http server")
	serveCmd.PersistentFlags().StringVar(&a.flags.serve.renderer, "renderer", "serve", "name of the default renderer to be used")
	rootCmd.AddCommand(serveCmd)

	// version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run:   a.versionCmd,
	}
	rootCmd.AddCommand(versionCmd)

	return a
}

func (a *App) renderCmd(cmd *cobra.Command, args []string) {
	cfg, err := NewConfig(a.flags.configfile)
	exitOnErr(err)

	// build system
	sys, errs := build(a.flags.base, a.flags.glob, a.flags.focus)
	exitOnErr(errs...)

	// render template
	renderer, ok := cfg.Renderer[a.flags.render.renderer]
	if !ok {
		exitOnErr(fmt.Errorf("renderer %s not specified in %s", a.flags.render.renderer, a.flags.configfile))
	}
	out, err := render(sys, renderer)
	exitOnErr(err)

	fmt.Println(out)
}

func (a *App) svgCmd(cmd *cobra.Command, args []string) {
	cfg, err := NewConfig(a.flags.configfile)
	exitOnErr(err)

	// build system
	sys, errs := build(a.flags.base, a.flags.glob, a.flags.focus)
	exitOnErr(errs...)

	// render template
	renderer, ok := cfg.Renderer[a.flags.svg.renderer]
	if !ok {
		exitOnErr(fmt.Errorf("renderer %s not specified in %s", a.flags.svg.renderer, a.flags.configfile))
	}
	out, err := render(sys, renderer)
	exitOnErr(err)

	// create svg
	img, err := svg(out)
	exitOnErr(err)

	err = os.WriteFile(a.flags.svg.out, img, 0644)
	exitOnErr(err)

	fmt.Printf("file '%s' written...\n", a.flags.svg.out)
}

func (a *App) serveCmd(cmd *cobra.Command, args []string) {
	s, err := NewServer(
		a.flags.serve.listener,
		a.flags.serve.renderer,
		a.flags.configfile,
		a.flags.base,
		a.flags.glob,
	)
	exitOnErr(err)
	err = s.Run()
	exitOnErr(err)
}

func (a *App) versionCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Version:   ", versioninfo.Version)
	fmt.Println("Revision:  ", versioninfo.Revision)
	fmt.Println("DirtyBuild:", versioninfo.DirtyBuild)
	fmt.Println("LastCommit:", versioninfo.LastCommit)
}

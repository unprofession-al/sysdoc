package main

import (
	"fmt"
	"os"
	"sysdoc/internal/persistence"

	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

type App struct {
	flags struct {
		configfile string
		base       string
		glob       string
		focus      []string
		git        struct {
			url     string
			user    string
			branch  string
			keyfile string
			pass    string
		}
		render struct {
			renderer      string
			out           string
			noPostprocess bool
		}
		serve struct {
			renderer     string
			listener     string
			cacheTimeout string
		}
	}

	// Postprocessors
	renderer *Renderer

	// entry point
	Execute func() error
}

func NewApp() *App {
	a := &App{}
	appName := "sysdoc"

	a.renderer = NewRenderer()

	// root
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: "sysdoc allows to document dependencies between systems",
	}
	rootCmd.PersistentFlags().StringVar(&a.flags.configfile, "config", "sysdoc.yaml", "configuration file path relative to the base")
	rootCmd.PersistentFlags().StringVar(&a.flags.base, "base", ".", "base directory of the sysdoc definitions")
	rootCmd.PersistentFlags().StringVar(&a.flags.glob, "glob", "README.md", "glob to find sysdoc definitions")
	rootCmd.PersistentFlags().StringSliceVar(&a.flags.focus, "focus", []string{}, "elements to be focussed")
	rootCmd.PersistentFlags().StringVar(&a.flags.git.url, "git.url", "", "url of git repo")
	rootCmd.PersistentFlags().StringVar(&a.flags.git.user, "git.user", os.Getenv("GIT_USER"), "git user name (can be set via environment variable 'GIT_USER')")
	rootCmd.PersistentFlags().StringVar(&a.flags.git.branch, "git.branch", "refs/heads/master", "git branch to be used")
	rootCmd.PersistentFlags().StringVar(&a.flags.git.pass, "git.pass", os.Getenv("GIT_PASS"), "pass for key file or git user (can be set via environment variable 'GIT_PASS')")
	rootCmd.PersistentFlags().StringVar(&a.flags.git.keyfile, "git.keyfile", "", "path to ssh key file")
	a.Execute = rootCmd.Execute

	// render
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "renders system documentation in a given template",
		Run:   a.renderCmd,
	}
	renderCmd.PersistentFlags().StringVar(&a.flags.render.renderer, "renderer", "default", "name of the renederer (set of templates in the configuration file)")
	renderCmd.PersistentFlags().StringVar(&a.flags.render.out, "out", "", "name of the file to be written (leave empty for STDOUT)")
	renderCmd.PersistentFlags().BoolVar(&a.flags.render.noPostprocess, "no-postprocess", false, "do not run post processor")
	rootCmd.AddCommand(renderCmd)

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
	serveCmd.PersistentFlags().StringVar(&a.flags.serve.cacheTimeout, "cache-timeout", "10m", "timeout of the internal cache")
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
	var pc persistence.Config
	pc.Filepath = a.flags.base
	pc.Git.URL = a.flags.git.url
	pc.Git.User = a.flags.git.user
	pc.Git.Pass = a.flags.git.pass
	pc.Git.Keyfile = a.flags.git.keyfile
	p, err := persistence.New(pc)
	exitOnErr(err)
	err = p.CheckoutBranch(a.flags.git.branch)
	exitOnErr(err)

	cfg, err := NewConfig(a.flags.configfile, p.Filesystem())
	exitOnErr(err)

	// build system
	sys, errs := NewSystem(a.flags.base, a.flags.glob, a.flags.focus, p)
	exitOnErr(errs...)

	// render template
	renderer, ok := cfg.Renderer[a.flags.render.renderer]
	if !ok {
		exitOnErr(fmt.Errorf("renderer %s not specified in %s", a.flags.render.renderer, a.flags.configfile))
	}
	data, err := a.renderer.Do(sys, renderer, a.flags.render.noPostprocess)
	exitOnErr(err)

	if a.flags.render.out != "" {
		err = os.WriteFile(a.flags.render.out, data, 0644)
		exitOnErr(err)
	} else {
		fmt.Println(string(data))
	}
}

func (a *App) serveCmd(cmd *cobra.Command, args []string) {
	var pc persistence.Config
	pc.Filepath = a.flags.base
	pc.Git.URL = a.flags.git.url
	pc.Git.User = a.flags.git.user
	pc.Git.Pass = a.flags.git.pass
	pc.Git.Keyfile = a.flags.git.keyfile
	p, err := persistence.New(pc)
	exitOnErr(err)
	err = p.CheckoutBranch(a.flags.git.branch)
	exitOnErr(err)

	s, err := NewServer(
		a.flags.serve.listener,
		a.flags.base,
		a.flags.glob,
		a.flags.serve.cacheTimeout,
		p,
		*a.renderer,
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

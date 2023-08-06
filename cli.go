package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

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
	tmpl, err := template.New("test").Parse(htmlTemplate)
	exitOnErr(err)

	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		focusElems := r.URL.Query().Get("focus")
		focus := strings.Split(focusElems, "_")

		rendererName := r.URL.Query().Get("renderer")
		if rendererName == "" {
			rendererName = a.flags.serve.renderer
		}

		cfg, err := NewConfig(a.flags.configfile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// build system
		sys, errs := build(a.flags.base, a.flags.glob, focus)
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
			exitOnErr(fmt.Errorf("renderer %s not specified in %s", rendererName, a.flags.configfile))
		}
		out, err := render(sys, renderer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// create svg
		img, err := svg(out)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var buf bytes.Buffer

		thing := strings.TrimSuffix(strings.TrimPrefix(string(img), `<?xml version="1.0" encoding="utf-8"?><svg`), "</xml>")
		thing = `<svg id="svg" class="svg"` + thing
		err = tmpl.Execute(&buf, thing)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(buf.Bytes())
	})

	http.HandleFunc("/svg/", func(w http.ResponseWriter, r *http.Request) {
		var focus []string
		for _, e := range strings.Split(r.URL.Path, "+") {
			focus = append(focus, strings.Trim(e, "/"))
		}

		rendererName := r.URL.Query().Get("renderer")
		if rendererName == "" {
			rendererName = a.flags.serve.renderer
		}

		cfg, err := NewConfig(a.flags.configfile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// build system
		sys, errs := build(a.flags.base, a.flags.glob, focus)
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
			exitOnErr(fmt.Errorf("renderer %s not specified in %s", rendererName, a.flags.configfile))
		}
		out, err := render(sys, renderer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// create svg
		img, err := svg(out)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(img)
	})

	fmt.Printf("server listening on http://%s/, hit CTRL-C to stop server...\n", a.flags.serve.listener)
	_ = http.ListenAndServe(a.flags.serve.listener, nil)
}

func (a *App) versionCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Version:   ", versioninfo.Version)
	fmt.Println("Revision:  ", versioninfo.Revision)
	fmt.Println("DirtyBuild:", versioninfo.DirtyBuild)
	fmt.Println("LastCommit:", versioninfo.LastCommit)
}

var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>My Website</title>

<style>
.svg {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  cursor: move;
}

.svg-scrim {
  pointer-events: none;
  z-index: 5;
}

.proxy {
  fill: none;
  stroke: none;
}

.controls {
  position: fixed;
  top: 0;
  left: 0;
  padding: 12px;
  z-index: 10;
}
.controls .controls-button {
  font-weight: 700;
}

.info {
  -webkit-user-select: none;
     -moz-user-select: none;
      -ms-user-select: none;
          user-select: none;
  pointer-events: none;
}
.info ul {
  font-size: 13px;
  list-style-type: none;
  padding: 0;
  line-height: 20px;
  margin-top: 0;
}

.svg-background {
  fill: none;
  stroke: none;
}

.pivot {
  fill: #ffc107;
  stroke: rgba(0, 0, 0, 0.5);
  stroke-width: 2;
  opacity: 0;
}


</style>



  </head>
  <body>

{{.}}











<svg id="svg-scrim" class="svg svg-scrim">  
  <circle id="pivot" class="pivot" cx="0" cy="0" r="6" />
</svg>

<div class="controls">
  <button id="reset">
    Reset
  </button>
</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/gsap/1.20.3/TweenMax.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/gsap/1.20.3/utils/Draggable.min.js"></script>

<script type="text/javascript">

// console.clear();

var svg = document.querySelector("#svg");
var reset = document.querySelector("#reset");
var pivot = document.querySelector("#pivot");
var proxy = document.createElement("div");
var viewport = document.querySelector("#d2-svg");

var rotateThreshold = 4;
var reachedThreshold = false;

var point = svg.createSVGPoint();
var startClient = svg.createSVGPoint();
var startGlobal = svg.createSVGPoint();

var viewBox = svg.viewBox.baseVal;

var cachedViewBox = {
  x: viewBox.x,
  y: viewBox.y,
  width: viewBox.width,
  height: viewBox.height
};

var zoom = {
  animation: new TimelineLite(),
  scaleFactor: 1.6,
  duration: 0.5,
  ease: Power2.easeOut,
};

TweenLite.set(pivot, { scale: 0 });

var resetAnimation = new TimelineLite();
var pivotAnimation = TweenLite.to(pivot, 0.1, {
  alpha: 1,
  scale: 1,
  paused: true,
});

var pannable = new Draggable(proxy, {
  throwResistance: 3000,
  trigger: svg,
  throwProps: true,
  onPress: selectDraggable,
  onDrag: updateViewBox,
  onThrowUpdate: updateViewBox,
});

var rotatable = new Draggable(viewport, {
  type: "rotation",
  trigger: svg,
  throwProps: true,
  liveSnap: true,
  snap: checkThreshold,
  onPress: selectDraggable,
});

rotatable.disable();


// TODO, Does not work
var connections = document.querySelectorAll(".connection");
connections.forEach((conn) => {
	console.log(conn)
conn.addEventListener("click", print);
});



reset.addEventListener("click", resetViewport);
window.addEventListener("wheel", onWheel);
window.addEventListener("resize", function() {
  pivotAnimation.reverse();
});

function print(event) {
	console.log("bla");
};

//window.addEventListener("contextmenu", function(event) {
//  event.preventDefault();
//	event.stopPropagation();
//  return false;
//});

//
// ON WHEEL
// =========================================================================== 
function onWheel(event) {
  // event.preventDefault();
  
  pivotAnimation.reverse();
  
  var normalized;  
  var delta = event.wheelDelta;

  if (delta) {
    normalized = (delta % 120) == 0 ? delta / 120 : delta / 12;
  } else {
    delta = event.deltaY || event.detail || 0;
    normalized = -(delta % 3 ? delta * 10 : delta / 3);
  }
  
  var scaleDelta = normalized > 0 ? 1 / zoom.scaleFactor : zoom.scaleFactor;
  
  point.x = event.clientX;
  point.y = event.clientY;
  
  var startPoint = point.matrixTransform(svg.getScreenCTM().inverse());
    
  var fromVars = {
    ease: zoom.ease,
    x: viewBox.x,
    y: viewBox.y,
    width: viewBox.width,
    height: viewBox.height,
  };
  
  viewBox.x -= (startPoint.x - viewBox.x) * (scaleDelta - 1);
  viewBox.y -= (startPoint.y - viewBox.y) * (scaleDelta - 1);
  viewBox.width *= scaleDelta;
  viewBox.height *= scaleDelta;
    
  zoom.animation = TweenLite.from(viewBox, zoom.duration, fromVars);  
}

//
// SELECT DRAGGABLE
// =========================================================================== 
function selectDraggable(event) {
   
  if (resetAnimation.isActive()) {
    resetAnimation.kill();
  }
    
  startClient.x = this.pointerX;
  startClient.y = this.pointerY;
  startGlobal = startClient.matrixTransform(svg.getScreenCTM().inverse());
  
  // Right mouse button
  if (event.button === 2) {
    
    reachedThreshold = false;
    
    TweenLite.set(viewport, { 
      svgOrigin: startGlobal.x + " " + startGlobal.y
    });
    
    TweenLite.set(pivot, { 
      x: this.pointerX, 
      y: this.pointerY
    });
       
    pannable.disable();
    rotatable.enable().update().startDrag(event);
    pivotAnimation.play(0);
    
  } else {
    
    TweenLite.set(proxy, { 
      x: this.pointerX, 
      y: this.pointerY
    });
    
    rotatable.disable();
    pannable.enable().update().startDrag(event);
    pivotAnimation.reverse();
  }
}

//
// RESET VIEWPORT
// =========================================================================== 
function resetViewport() {
    
  var duration = 0.8;
  var ease = Power3.easeOut;
  
  pivotAnimation.reverse();
  
  if (pannable.tween) {
    pannable.tween.kill();
  }
  
  if (rotatable.tween) {
    rotatable.tween.kill();
  }
    
  resetAnimation.clear()
    .to(viewBox, duration, {
      x: cachedViewBox.x,
      y: cachedViewBox.y,
      width: cachedViewBox.width,
      height: cachedViewBox.height,
      ease: ease
    }, 0)
    .to(viewport, duration, {
      attr: { transform: "matrix(1,0,0,1,0,0)" },
      // rotation: "0_short",
      smoothOrigin: false,
      svgOrigin: "0 0",
      ease: ease
    }, 0)
}

//
// CHECK THRESHOLD
// =========================================================================== 
function checkThreshold(value) {
  
  if (reachedThreshold) {
    return value;
  }
  
  var dx = Math.abs(this.pointerX - startClient.x);
  var dy = Math.abs(this.pointerY - startClient.y);
  
  if (dx > rotateThreshold || dy > rotateThreshold || this.isThrowing) {
    reachedThreshold = true;
    return value;
  }
    
  return this.rotation;
}

//
// UPDATE VIEWBOX
// =========================================================================== 
function updateViewBox() {
  
  if (zoom.animation.isActive()) {
    return;
  }
  
  point.x = this.x;
  point.y = this.y;
  
  var moveGlobal = point.matrixTransform(svg.getScreenCTM().inverse());
    
  viewBox.x -= (moveGlobal.x - startGlobal.x);
  viewBox.y -= (moveGlobal.y - startGlobal.y); 
}



</script>

  </body>
</html>

`

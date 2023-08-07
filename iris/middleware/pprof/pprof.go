// Package pprof provides native pprof support via middleware. See _examples/miscellaneous/pprof
package pprof

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/handlerconv"
	"html/template"
	"net/http/pprof"
	rpprof "runtime/pprof"
	"sort"
	"strings"
)

func init() {
	context.SetHandlerName("go-utils/iris/middleware/pprof.*", "go-utils.pprof")
}

// New returns a new pprof (profile, cmdline, symbol, goroutine, heap, threadcreate, debug/block) Middleware.
// Note: Route MUST have the last named parameter wildcard named '{action:path}'
func New() context.Handler {
	//indexHandler := handlerconv.FromStd(pprof.Index)
	cmdlineHandler := handlerconv.FromStd(pprof.Cmdline)
	profileHandler := handlerconv.FromStd(pprof.Profile)
	traceHandler := handlerconv.FromStd(pprof.Trace)

	return func(ctx iris.Context) {
		action := ctx.Params().Get("action")
		switch action {
		case "cmdline":
			cmdlineHandler(ctx)
		case "profile":
			profileHandler(ctx)
		case "trace":
			traceHandler(ctx)
		default:
			pprofIndex(ctx)
		}
	}
}

var profileDescriptions = map[string]string{
	"allocs":       "A sampling of all past memory allocations",
	"block":        "Stack traces that led to blocking on synchronization primitives",
	"cmdline":      "The command line invocation of the current program",
	"goroutine":    "Stack traces of all current goroutines",
	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":        "Stack traces of holders of contended mutexes",
	"profile":      "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
	"trace":        "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
}

func pprofIndex(ctx iris.Context) {
	if strings.HasPrefix(ctx.Path(), "/debug/pprof/") {
		name := strings.TrimPrefix(ctx.Path(), "/debug/pprof/")
		if name != "" {
			handlerconv.FromStd(pprof.Handler(name))(ctx)
			return
		}
	}

	type profile struct {
		Name  string
		Href  string
		Desc  string
		Count int
	}
	var profiles []profile
	for _, p := range rpprof.Profiles() {
		profiles = append(profiles, profile{
			Name:  p.Name(),
			Href:  p.Name() + "?debug=1",
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}

	// Adding other profiles exposed from within this package
	for _, p := range []string{"cmdline", "profile", "trace"} {
		profiles = append(profiles, profile{
			Name: p,
			Href: p,
			Desc: profileDescriptions[p],
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	data := map[string]interface{}{
		"Profiles": profiles,
		"Path":     ctx.RequestPath(false),
	}
	if err := indexTmpl.Execute(ctx, data); err != nil {
		ctx.Application().Logger().Error(err)
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>{{.Path}}</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
{{.Path}}<br>
<br>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td></thead>
{{$path := .Path}}
{{range .Profiles}}
	<tr>
	<td>{{.Count}}</td><td><a href="{{$path}}/{{.Href}}">{{.Name}}</a></td>
	</tr>
{{end}}
</table>
<a href="{{$path}}/goroutine?debug=2">full goroutine stack dump</a>
<br/>
<p>
Profile Descriptions:
<ul>
{{range .Profiles}}
<li><div class=profile-name>{{.Name}}:</div> {{.Desc}}</li>
{{end}}
</ul>
</p>
</body>
</html>
`))

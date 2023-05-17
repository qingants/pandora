package minirpc

import (
	"fmt"
	"html/template"
	"net/http"
)

const debugPage = `
<html>
	<body>
	<title>MiniRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
</html>
`

var t = template.Must(template.New("mini rpc debug").Parse(debugPage))

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

// Runs at /debug/minirpc
func (s debugHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Build a sorted version of the data
	var services []debugService
	s.serviceMap.Range(func(key, value any) bool {
		svc := value.(*service)
		services = append(services, debugService{
			Name:   key.(string),
			Method: svc.method,
		})
		return true
	})
	err := t.Execute(w, services)
	if err != nil {
		_, _ = fmt.Fprintln(w, "mini rpc: error executing template: ", err.Error())
	}
}

// debug/minirpc
// func (s *Server) HandleHTTP() {
// 	http.Handle(defaultRPCPath, s)
// 	http.Handle(defaultDebugPath, debugHTTP{s})
// 	log.Println("mini rpc server debug path: ", defaultDebugPath)
// }



{{range $index, $el := .Items}}
func init{{$el.Name}}() {{$el.Type}} {
    {{$idx, $dep := range $el.Deps}}
        {{$dep.Name}} := init{{$dep.Name}}()
    {{end}}
    {{$el.Name}}, err := template factory...???}}
}
{{end}}
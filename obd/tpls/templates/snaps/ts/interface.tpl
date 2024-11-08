export interface {{.name}} {
{{ range .fields }}  {{index . 0}}: {{index . 1}}
{{end}}}

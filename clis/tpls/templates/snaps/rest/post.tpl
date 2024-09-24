import strings
import github.com/google/uuid

data.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
{{ range .fields }}
{{if .is_pointer}}
if opts.{{.name}} != nil {
    data.{{.name}} = *opts.{{.name}}
}
{{else}}
data.{{.name}} = opts.{{.name}}
{{end}}

{{ end }}

err = cfg.DB().Create(data).Error

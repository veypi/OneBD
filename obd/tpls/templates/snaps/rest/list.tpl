query := cfg.DB()

{{ range .fields }}
{{if .is_pointer}}
if opts.{{.name}} != nil {
    {{if eq .type  "string"}}
    query = query.Where("{{.snake}} LIKE ?", opts.{{.name}})
    {{else if eq .type "&{time Time}"}}
    query = query.Where("{{.snake}} > ?", opts.{{.name}})
    {{else}}
    query = query.Where("{{.snake}} = ?", opts.{{.name}})
    {{end}}
}
{{else}}
{{if eq .type  "string"}}query = query.Where("{{.snake}} LIKE ?", opts.{{.name}})
{{else if eq .type "&{time Time}"}}query = query.Where("{{.snake}} > ?", opts.{{.name}})
{{else}}query = query.Where("{{.snake}} = ?", opts.{{.name}}){{end}}
{{end}}
{{end}}

err = query.Find(&data).Error

query := cfg.DB()

{{ range .fields }}
{{if eq .type  "string"}}query = query.Where("{{.snake}} LIKE ?", opts.{{.name}})
{{else if eq .type "&{time Time}"}}query = query.Where("{{.snake}} > ?", opts.{{.name}})
{{end}}{{end}}

err = query.Find(&data).Error

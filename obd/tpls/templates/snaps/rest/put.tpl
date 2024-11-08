err = cfg.DB().Where("id = ?", opts.ID).First(data).Error
if err != nil {
	return nil, err
}
optsMap := map[string]interface{}{
{{ range .fields }}
    "{{.snake}}": opts.{{.name}},
{{end}}
}
err = cfg.DB().Model(data).Updates(optsMap).Error

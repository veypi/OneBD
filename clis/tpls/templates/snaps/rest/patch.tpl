err = cfg.DB().Where("id = ?", opts.ID).First(data).Error
if err != nil {
	return nil, err
}
optsMap := make(map[string]interface{})
{{ range .fields }}
{{if .is_pointer}}
if opts.{{.name}} != nil {
	optsMap["{{.snake}}"] = opts.{{.name}}
}
{{end}}
{{end}}
err = cfg.DB().Model(data).Updates(optsMap).Error

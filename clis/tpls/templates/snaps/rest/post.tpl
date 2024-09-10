import strings
import github.com/google/uuid

data.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
{{ range .fields }}
data.{{.name}} = opts.{{.name}}{{ end }}

err = cfg.DB().Create(data).Error

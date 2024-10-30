export function {{.name}}({{.args}}) {
  return webapi.{{.method}}<{{if .is_list}}models.{{.obj}}[]{{else}}models.{{.obj}}{{end}}>(`{{.url}}`, { {{.resp}} })
}

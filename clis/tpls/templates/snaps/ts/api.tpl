export function {{.name}}({{.args}}) {
  return webapi.{{.method}}<models.{{.obj}}>(`{{.url}}`, { {{.resp}} })
}

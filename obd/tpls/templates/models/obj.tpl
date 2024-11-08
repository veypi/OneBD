//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package {{.package}}

{{if not .isRoot}}
import (
    "{{.common.repo}}/{{.common.model}}"
)
{{end}}

type {{.Obj}} struct {
    {{if .isRoot}}BaseModel{{else}}{{.common.model}}.BaseModel{{end}}
}

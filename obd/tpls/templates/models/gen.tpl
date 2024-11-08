//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//
{{.common.noedit}}

package {{.package}}

import (
    {{range .imports}}"{{.}}"{{end}}
)

{{.body}}

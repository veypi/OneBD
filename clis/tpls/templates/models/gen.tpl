//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//
{{.common.noedit}}

package models

import (
    {{range .imports}}"{{.}}"{{end}}
)

{{.structs}}

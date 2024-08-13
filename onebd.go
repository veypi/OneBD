//
// onebd.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-08 19:41
// Distributed under terms of the MIT license.
//

package OneBD

import "github.com/veypi/OneBD/rest/router"

const (
	Version = "v0.5.0"
)

type X = router.X

var NewRouter = router.NewRouter

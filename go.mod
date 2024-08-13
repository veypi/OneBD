module github.com/veypi/OneBD

go 1.22.5

require (
	github.com/veypi/utils v0.3.7
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
)

require (
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/rs/zerolog v1.17.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/veypi/utils => ../OceanCurrent/utils/

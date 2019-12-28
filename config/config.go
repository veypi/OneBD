package config

type Config struct {
	// 服务监听地址
	Host          string
	Charset       string `json:"charset,omitempty"`
	TimeFormat    string `json:"time_format,omitempty"`
	PostMaxMemory int64
}

func (c *Config) IsValid() *Config {
	if c.Host == "" {
		c.Host = "0.0.0.0:8000"
	}
	return c
}

func DefaultConfig() *Config {
	c := &Config{}
	return c.IsValid()
}

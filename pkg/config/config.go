package config

type Config struct {
	Poll                 int    `env:"NTPC_POLL" envDefault:"5"`
	RemoteHost           string `env:"NTPC_REMOTE_HOST" envDefault:"time.google.com"`
	RemotePort           int    `env:"NTPC_REMOTE_PORT" envDefault:"123"`
	SystimeUpdateEnabled bool   `env:"NTPC_SYSTIME_UPDATE_ENABLED" envDefault:"true"`
}

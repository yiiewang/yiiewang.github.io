package config

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type ServerConfig struct {
	Name        string        `mapstructure:"name"`
	Port        string        `mapstructure:"port"`
	UserSrvInfo UserSrvConfig `mapstructure:"user_srv"`
}

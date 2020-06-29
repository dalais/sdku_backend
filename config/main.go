package config

// LocalConfig ...
type LocalConfig struct {
	APPKey    string `yaml:"app_key"`
	DebugMode bool   `yaml:"debug_mode"`
	Server    struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Connection string `yaml:"connection"`
		Host       string `yaml:"host"`
		Port       string `yaml:"port"`
		Db         string `yaml:"database"`
		User       string `yaml:"user"`
		Pass       string `yaml:"pass"`
	} `yaml:"database"`
}

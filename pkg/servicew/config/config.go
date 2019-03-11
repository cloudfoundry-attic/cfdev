package config

type Config struct {
	Args       []string
	Env        map[string]string
	Executable string
	Label      string
	Log        string
	Options    map[string]interface{}
}

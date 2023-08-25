package postprocessor

type Postprocessor interface {
	Process(string) ([]byte, error)
}

type Config struct {
	Name   string            `yaml:"name"`
	Params map[string]string `yaml:"params"`
}

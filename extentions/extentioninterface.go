package extentions

type Extentioninterface interface {
	Name() string
	Process(data string, param map[string]string) (string, error)
}

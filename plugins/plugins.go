package plugins

type Plugin interface {
	DefaultPlay() string
	Mask(string) string
	Unmask(string) string
}

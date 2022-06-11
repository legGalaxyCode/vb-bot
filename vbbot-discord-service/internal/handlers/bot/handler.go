package bot

type Handler interface {
	Handle() error
}

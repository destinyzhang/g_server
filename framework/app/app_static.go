package app

type IModule interface {
	Run()
	Start() bool
	Stop() bool
	Name() string
}

func NewApp() *App {
	return &App{modules: make([]IModule, 0, 10), stop: false}
}

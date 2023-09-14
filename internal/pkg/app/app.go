package app

type Application struct {
}

func New() Application {
	return Application{}
}

func (app Application) Run() int {
	return 1
}

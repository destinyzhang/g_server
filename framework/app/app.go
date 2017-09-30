package app

import (
	"g_server/framework/com"
	"time"
)

type App struct {
	modules []IModule
	stop    bool
	now     time.Time
}

func (app *App) RegModule(mod IModule) *App {
	app.modules = append(app.modules, mod)
	return app
}

func (app *App) GetModule(name string) IModule {
	for _, mod := range app.modules {
		if name == mod.Name() {
			return mod
		}
	}
	return nil
}

func (app *App) Now() time.Time {
	return app.now
}

func (app *App) Millisecond() int64 {
	return app.now.UnixNano() / int64(time.Millisecond)
}

func (app *App) Start() {
	for _, mod := range app.modules {
		if !mod.Start() {
			panic("mod start fail!!!!!" + mod.Name())
		}
	}
}

func (app *App) Stop() {
	app.stop = true
}
func (app *App) destory() {
	for _, mod := range app.modules {
		com.SafeCall(func() {
			mod.Stop()
		})
	}
}

func (app *App) Run() {
	for !app.stop {
		app.now = time.Now()
		for _, module := range app.modules {
			com.SafeCall(func() {
				module.Run()
			})
		}
		time.Sleep(time.Millisecond)
	}
	app.destory()
}

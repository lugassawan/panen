package presenter

import (
	"context"
	"fmt"
)

// App struct provides the application's bound methods.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so it can be used for runtime calls (dialogs, events, etc.).
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to Panen!", name)
}

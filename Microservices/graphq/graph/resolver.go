package graph

import (
	"lab1/app"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	application *app.App
}

func NewResolver(application *app.App) *Resolver {
	return &Resolver{application: application}
}

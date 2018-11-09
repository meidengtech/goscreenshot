package main

import (
	"context"
)

// IShot blabla
type IShot interface {
	Do(context.Context, string, int) ([]byte, error)
}

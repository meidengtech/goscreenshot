package main

import (
	"context"
)

// IShot blabla
type IShot interface {
	Do(context.Context, string, string, int) ([]byte, error)

	Stat() string
}

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	commonPool "github.com/jolestar/go-commons-pool"
)

// ChromeObject for Pooled
type ChromeObject struct {
	cdpRes *chromedp.Res
}

// Release releases the cdpRes
func (c *ChromeObject) Release() error {
	return c.cdpRes.Release()
}

// ChromeObjectFactory for Pooled
type ChromeObjectFactory struct {
	cdpPool    *chromedp.Pool
	chromePath string
}

// MakeObject make a new Object
func (f *ChromeObjectFactory) MakeObject(ctxt context.Context) (*commonPool.PooledObject, error) {
	log.Println("start new chrome")
	c, err := f.cdpPool.Allocate(ctxt,
		runner.Flag("headless", true),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-first-run", true),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-sandbox", true),
		runner.ExecPath(f.chromePath),
	)
	fmt.Println("starting new chrome now: ", c, err)
	if err != nil {
		return nil, err
	}
	return commonPool.NewPooledObject(&ChromeObject{cdpRes: c}), nil
}

// DestroyObject destroy an Object
func (f *ChromeObjectFactory) DestroyObject(ctxt context.Context, obj *commonPool.PooledObject) error {
	log.Println("stop new chrome")
	o := obj.Object.(*ChromeObject)
	o.Release()
	return nil
}

// ValidateObject check an Object valid
func (f *ChromeObjectFactory) ValidateObject(ctxt context.Context, o *commonPool.PooledObject) bool {
	//do validate
	return true
}

// ActivateObject make an object active
func (f *ChromeObjectFactory) ActivateObject(ctxt context.Context, o *commonPool.PooledObject) error {
	//do activate
	return nil
}

// PassivateObject make an object Passivate
func (f *ChromeObjectFactory) PassivateObject(ctxt context.Context, o *commonPool.PooledObject) error {
	//do passivate
	return nil
}

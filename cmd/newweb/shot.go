package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	commonPool "github.com/jolestar/go-commons-pool"
)

// IShot blabla
type IShot interface {
	Do(context.Context, string, int) ([]byte, error)
}

// EmptyShot blabla
type EmptyShot struct {
}

// Do blabla
func (e EmptyShot) Do(ctx context.Context, url string, width int) ([]byte, error) {
	s := "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="
	return base64.StdEncoding.DecodeString(s)
}

// PooledShotter is a PooledShotter Object
type PooledShotter struct {
	cPool *commonPool.ObjectPool
	log   *logrus.Logger
}

// NewPooledShotter chrome pool
func NewPooledShotter(cdpPool *chromedp.Pool) *PooledShotter {
	var err error
	if err != nil {
		log.Fatal(err)
	}
	p := PooledShotter{}
	cof := ChromeObjectFactory{cdpPool: cdpPool, chromePath: `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`}
	p.cPool = commonPool.NewObjectPoolWithDefaultConfig(context.TODO(), &cof)
	p.cPool.Config.MaxTotal = 10
	p.cPool.Config.MaxIdle = 1
	p.cPool.Config.MinIdle = 1
	p.cPool.Config.LIFO = false
	return &p
}

// Release all Chrome
func (p *PooledShotter) Release() {
	log.Println("do release now")
	p.cPool.Close(context.Background())
}

// Do with url and width
func (p *PooledShotter) Do(ctxt1 context.Context, url string, width int) ([]byte, error) {
	var picbuf []byte
	f := func() bool {
		cobj, err := p.cPool.BorrowObject(context.TODO())
		if err != nil {
			log.Panic(err)
		}

		defer p.cPool.ReturnObject(context.TODO(), cobj)
		if err != nil {
			log.Panic(err)
		}

		o := cobj.(*ChromeObject)
		c := o.cdpRes

		runctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		defer cancel()

		timeout := time.After(2 * time.Second)
		cc := make(chan bool)
		go func() {
			c.Run(runctx, screenshot(url, &picbuf, width))
			cc <- true
		}()

		select {
		case <-cc:
			fmt.Println("done")
			return true
		case <-timeout:
			return false
		}
	}
	for {
		if f() {
			break
		}
	}
	return picbuf, nil
}

func screenshot(urlstr string, picbuf *[]byte, width int) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		setViewportAndScale(int64(width), 1600, 1.0),
		chromedp.Sleep(50 * time.Millisecond),
		chromedp.WaitVisible("ImgLoadedFlagACHHcLIkD3", chromedp.ByID),
		chromedp.Screenshot("ACHHcLIkD3", picbuf, chromedp.NodeVisible, chromedp.ByID),
	}
}

func setViewportAndScale(w, h int64, scale float64) chromedp.ActionFunc {
	return func(ctxt context.Context, ha cdp.Executor) error {
		sw, sh := int64(float64(w)/scale), int64(float64(h)/scale)
		err := emulation.SetDeviceMetricsOverride(sw, sh, scale, false).WithScale(scale).Do(ctxt, ha)
		if err != nil {
			return err
		}
		return nil
	}
}

// PooledShotter2 is a PooledShotter Object
type PooledShotter2 struct {
	cPool *chromedp.Pool
	log   *logrus.Logger
}

// Do xxx
func (p *PooledShotter2) Do(ctxt1 context.Context, url string, width int) ([]byte, error) {
	var picbuf []byte
	c, err := p.cPool.Allocate(context.TODO(),
		runner.Flag("headless", true),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-first-run", true),
		runner.Flag("disable-gpu", true),
		// runner.Flag("no-sandbox", true),
	)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	c.Run(context.TODO(), screenshot(url, &picbuf, width))
	return picbuf, nil
}

// QueuedShotter xxx
type QueuedShotter struct {
	cPool *chromedp.Pool
}

// Start 启动dispatcher
func (p *QueuedShotter) Start() {
	StartDispatcher(4, p.cPool)
}

// Do xxx
func (p *QueuedShotter) Do(ctxt1 context.Context, url string, width int) ([]byte, error) {
	resp := make(chan WorkResponse)
	work := WorkRequest{Name: url, Response: resp}
	WorkQueue <- work
	response := <-resp
	return response.Picbuf, response.Error
}

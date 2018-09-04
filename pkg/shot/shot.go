package shot

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"

	"github.com/chromedp/chromedp/runner"
	commonPool "github.com/jolestar/go-commons-pool"
	"github.com/sempr/goscreenshot/constants"
)

func prepareChromeRes(ctxt context.Context, cPool *chromedp.Pool) (*chromedp.Res, error) {
	c, err := cPool.Allocate(ctxt,
		runner.Flag("headless", true),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-first-run", true),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-sandbox", true),
		runner.Flag("user-data-dir", constants.UserDataDir),
		runner.ExecPath(constants.ChromePath),
	)
	if err != nil {
		log.Println("prepareChromeRes: ", err)
	}
	return c, err
}

// Screenshot with url and width
func (p *PooledShotter) Screenshot(url string, width int) ([]byte, error) {
	var err error
	cobj, err := p.cPool.BorrowObject(p.ctx)
	if err != nil {
		log.Panic(err)
	}
	defer p.cPool.ReturnObject(p.ctx, cobj)

	o := cobj.(*ChromeObject)
	ctxt := o.Ctxt
	c := o.CdpRes
	var picbuf []byte
	err = c.Run(*ctxt, screenshot(url, &picbuf, width))
	if err != nil {
		return nil, err
	}
	return picbuf, nil
}

func screenshot(urlstr string, picbuf *[]byte, width int) chromedp.Tasks {
	fmt.Println(urlstr)
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		setViewportAndScale(int64(width), 1600, 1.0),
		chromedp.WaitVisible("ImgLoadedFlagACHHcLIkD3", chromedp.ByID),
		chromedp.Screenshot("ACHHcLIkD3", picbuf, chromedp.ByID),
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

// ChromeObject for Pooled
type ChromeObject struct {
	Ctxt   *context.Context
	CdpRes *chromedp.Res
	Count  int
}

// ChromeObjectFactory for Pooled
type ChromeObjectFactory struct {
	CdpPool *chromedp.Pool
}

// MakeObject make a new Object
func (f *ChromeObjectFactory) MakeObject(ctxt context.Context) (*commonPool.PooledObject, error) {
	log.Println("new chromeobject")
	c, err := prepareChromeRes(ctxt, f.CdpPool)
	if err != nil {
		return nil, err
	}
	return commonPool.NewPooledObject(&ChromeObject{Ctxt: &ctxt, CdpRes: c, Count: 0}), nil
}

// DestroyObject destroy an Object
func (f *ChromeObjectFactory) DestroyObject(ctxt context.Context, obj *commonPool.PooledObject) error {
	log.Println("destroy chromeobject")
	o := obj.Object.(*ChromeObject)
	o.CdpRes.Release()
	//do destroy
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

// PooledShotter is a PooledShotter Object
type PooledShotter struct {
	ctx       context.Context
	cPool     *commonPool.ObjectPool
	DebugMode bool
}

// Init chrome pool
func (p *PooledShotter) Init() {
	var cdpPool *chromedp.Pool
	var err error
	if p.DebugMode {
		cdpPool, err = chromedp.NewPool(chromedp.PoolLog(log.Printf, log.Printf, log.Printf), chromedp.PortRange(50100, 50199))
	} else {
		cdpPool, err = chromedp.NewPool(chromedp.PortRange(50100, 50199))
	}
	if err != nil {
		log.Fatal(err)
	}
	p.ctx = context.Background()
	cof := ChromeObjectFactory{CdpPool: cdpPool}
	p.cPool = commonPool.NewObjectPoolWithDefaultConfig(p.ctx, &cof)
	p.cPool.Config.MaxTotal = 3
}

// Release all Chrome
func (p *PooledShotter) Release() {
	fmt.Println("do release now")
	p.cPool.Close(p.ctx)
}

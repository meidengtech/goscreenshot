package shot

import (
	"context"
	"fmt"
	"log"
	"time"

	commonPool "github.com/jolestar/go-commons-pool"
	cdp "github.com/knq/chromedp"
	cdptypes "github.com/knq/chromedp/cdp"
	"github.com/knq/chromedp/cdp/emulation"
	"github.com/knq/chromedp/runner"
	"github.com/sempr/goscreenshot/constants"
)

var cPool *commonPool.ObjectPool

func prepareChromeRes(ctxt context.Context, cPool *cdp.Pool) (*cdp.Res, error) {
	c, err := cPool.Allocate(ctxt,
		runner.Flag("headless", true),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-first-run", true),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-sandbox", true),
		runner.Flag("user-data-dir", constants.UserDataDir),
		runner.ExecPath(constants.ChromePath),
	)
	return c, err
}

// Screenshot with url and width
func Screenshot(url string, width int) ([]byte, error) {
	var err error
	cobj, err := cPool.BorrowObject()
	if err != nil {
		log.Panic(err)
	}
	defer cPool.ReturnObject(cobj)
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

func screenshot(urlstr string, picbuf *[]byte, width int) cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(urlstr),
		setViewportAndScale(int64(width), 1600, 1.0),
		cdp.Sleep(50 * time.Millisecond),
		cdp.WaitVisible("#ACHHcLIkD3", cdp.ByQuery),
		cdp.Screenshot("#ACHHcLIkD3", picbuf, cdp.ByQuery),
	}
}

func setViewportAndScale(w, h int64, scale float64) cdp.ActionFunc {
	return func(ctxt context.Context, ha cdptypes.Handler) error {
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
	Ctxt       *context.Context
	CancleFunc *context.CancelFunc
	CdpRes     *cdp.Res
	Count      int
}

// ChromeObjectFactory for Pooled
type ChromeObjectFactory struct {
	CdpPool *cdp.Pool
}

func (f *ChromeObjectFactory) MakeObject() (*commonPool.PooledObject, error) {
	ctxt, cancel := context.WithCancel(context.Background())
	c, err := prepareChromeRes(ctxt, f.CdpPool)
	if err != nil {
		cancel()
		return nil, err
	}
	return commonPool.NewPooledObject(&ChromeObject{Ctxt: &ctxt, CancleFunc: &cancel, CdpRes: c, Count: 0}), nil
}

func (f *ChromeObjectFactory) DestroyObject(obj *commonPool.PooledObject) error {
	fmt.Println("destroy chromeobject")
	o := obj.Object.(*ChromeObject)
	o.CdpRes.Release()
	(*o.CancleFunc)()
	//do destroy
	return nil
}

func (f *ChromeObjectFactory) ValidateObject(o *commonPool.PooledObject) bool {
	//do validate
	return true
}

func (f *ChromeObjectFactory) ActivateObject(o *commonPool.PooledObject) error {
	//do activate
	return nil
}

func (f *ChromeObjectFactory) PassivateObject(o *commonPool.PooledObject) error {
	//do passivate
	return nil
}

func Init() {
	cdpPool, err := cdp.NewPool( /* cdp.PoolLog(log.Printf, log.Printf, log.Printf),cdp.PortRange(6000, 6005) */ )
	if err != nil {
		log.Fatal(err)
	}
	cof := ChromeObjectFactory{CdpPool: cdpPool}
	cPool = commonPool.NewObjectPoolWithDefaultConfig(&cof)
	cPool.Config.MaxTotal = 4
}

// Release all Chrome
func Release() {
	fmt.Println("do release now")
	cPool.Close()
}

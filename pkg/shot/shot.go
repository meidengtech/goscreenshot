package shot

import (
	"context"
	"fmt"
	"log"
	"time"

	cdp "github.com/knq/chromedp"
	cdptypes "github.com/knq/chromedp/cdp"
	"github.com/knq/chromedp/cdp/emulation"
	"github.com/knq/chromedp/runner"
)

var pool *cdp.Pool

// Screenshot with url and width
func Screenshot(url string, width int) ([]byte, error) {
	var err error
	fmt.Println(url, width)
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := pool.Allocate(ctxt,
		runner.Flag("headless", true),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-first-run", true),
		runner.ExecPath(`/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary`),
	)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	if err != nil {
		return nil, err
	}

	var picbuf []byte
	err = c.Run(ctxt, screenshot(url, &picbuf, width))
	if err != nil {
		return nil, err
	}
	return picbuf, nil
}

func screenshot(urlstr string, picbuf *[]byte, width int) cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(urlstr),
		setViewportAndScale(int64(width), 600, 1.0),
		cdp.Sleep(100 * time.Millisecond),
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

func init() {
	var err error
	pool, err = cdp.NewPool(
	// cdp.PoolLog(log.Printf, log.Printf, log.Printf),
	// cdp.PortRange(6000, 6005),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// Release all Chrome
func Release() {
	pool.Shutdown()
}

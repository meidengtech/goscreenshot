package main

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// IShot blabla
type IShot interface {
	Do(context.Context, string, int) ([]byte, error)
}

func screenshot(urlstr string, picbuf *[]byte, width int) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		setViewportAndScale(int64(width), 1600, 1.0),
		chromedp.Sleep(10 * time.Millisecond),
		// chromedp.WaitVisible("ImgLoadedFlagACHHcLIkD3", chromedp.ByID),
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

package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

type WorkRequest struct {
	Name     string
	Width    int
	Response chan WorkResponse
}

type WorkResponse struct {
	Picbuf []byte
	Error  error
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
	Pt          *devtool.Target
}

func NewWorker(id int, workerQueue chan chan WorkRequest, pt *devtool.Target) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		Pt:          pt,
	}

	return worker
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case work := <-w.Work:
				// Receive a work request.
				log.Printf("worker%d: Hello, %s!\n", w.ID, work.Name)
				var picbuf []byte

				picbuf, err := w.doScreenShot(context.TODO(), w.Pt, work.Name, work.Width)
				if err != nil {
					logrus.Panic(err)
				}

				wr := WorkResponse{picbuf, nil}
				work.Response <- wr

			case <-w.QuitChan:
				// We have been asked to stop.
				log.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.QuitChan <- true
}

var WorkerQueue chan chan WorkRequest
var WorkQueue = make(chan WorkRequest, 100)

func (p *QueuedShotter) StartDispatcher(nworkers int) {
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	log.Println(p.chromeServer)
	devt := devtool.New(p.chromeServer)
	p.devt = devt
	for i := 0; i < nworkers; i++ {
		log.Println("Starting worker", i+1)
		pt, err := devt.CreateURL(context.TODO(), "about:blank")
		if err != nil {
			log.Panic(err)
		}
		worker := NewWorker(i+1, WorkerQueue, pt)
		worker.Start()
		p.workers = append(p.workers, worker)
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				log.Println("Received work requeust")
				go func() {
					worker := <-WorkerQueue

					log.Println("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

// QueuedShotter xxx
type QueuedShotter struct {
	devt         *devtool.DevTools
	workers      []Worker
	chromeServer string
	log          *logrus.Logger
}

func (p *QueuedShotter) Stop() {
	var wg sync.WaitGroup
	for _, w := range p.workers {
		log.Println("stop worker: ", w.ID)
		wg.Add(1)
		go func(nw Worker) {
			log.Println("stopping worker: ", nw.ID)
			p.devt.Close(context.TODO(), nw.Pt)
			wg.Done()
			log.Println("stopped worker: ", nw.ID)
		}(w)
	}
	wg.Wait()
}

// Do xxx
func (p *QueuedShotter) Do(ctxt1 context.Context, url string, width int) ([]byte, error) {
	resp := make(chan WorkResponse)
	work := WorkRequest{Name: url, Response: resp, Width: width}
	WorkQueue <- work
	response := <-resp
	return response.Picbuf, response.Error
}

func (w *Worker) doScreenShot(ctx context.Context, pt *devtool.Target, url string, width int) ([]byte, error) {
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)
	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		return nil, err
	}

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs(url)
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		return nil, err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		return nil, err
	}

	logrus.Infof("Page loaded with frame ID: %s\n", nav.FrameID)
	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	getvp := func() *page.Viewport {
		doc2, _ := c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{NodeID: doc.Root.NodeID, Selector: "#ACHHcLIkD3"})
		c.DOM.SetAttributeValue(ctx, &dom.SetAttributeValueArgs{NodeID: doc2.NodeID, Name: "style", Value: fmt.Sprintf("width: %dpx", width)})
		rect, _ := c.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{NodeID: &doc2.NodeID})

		vp := page.Viewport{
			X:      rect.Model.Content[0],
			Y:      rect.Model.Content[1],
			Width:  float64(rect.Model.Width),
			Height: float64(rect.Model.Height),
			Scale:  1.0,
		}
		return &vp
	}
	step := 20
	var vp *page.Viewport
	for t := 0; t < 3000; t += step {
		vp = getvp()
		logrus.Println(vp)
		if vp.Y == 0 {
			break
		}
	}
	log.Println(vp)
	// Capture a screenshot of the current page.
	screenshotArgs := page.NewCaptureScreenshotArgs().
		SetClip(*vp).
		SetFormat("jpeg").SetQuality(80)
	screenshot, err := c.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		return nil, err
	}
	return screenshot.Data, err
}

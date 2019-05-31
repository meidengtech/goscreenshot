package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

type WorkInfo struct {
    TimeCreate string
    TimeStart  string
    TimeEnd    string
    Logs       []string
}

type WorkRequest struct {
	Name     string
	Html     string
	Width    int
	Response chan WorkResponse
	Ctx      context.Context
	Info     WorkInfo
}

func NewWorkRequest(id string, response chan WorkResponse, html string, width int, ctx context.Context) WorkRequest {
	request := WorkRequest{
		Name:     id,
		Html:     html,
		Width:    width,
		Response: response,
		Ctx:      ctx,
	}
	request.Info.TimeCreate = time.Now().String()

	return request
}

type WorkResponse struct {
	Picbuf []byte
	Error  error
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan *Worker
	QuitChan    chan bool
	Pt          *devtool.Target
	LastWork    WorkRequest
}

func NewWorker(id int, workerQueue chan *Worker, pt *devtool.Target) Worker {
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

func (w *Worker) Json() string {
    var builder strings.Builder
    fmt.Fprintln(&builder, "\t{")
    fmt.Fprintf(&builder, "\t\t\"ID\": %d,\n", w.ID)
    fmt.Fprintf(&builder, "\t\t\"LastWork\": {\n")
    fmt.Fprintf(&builder, "\t\t\t\"Name\": \"%s\",\n", w.LastWork.Name)
    fmt.Fprintf(&builder, "\t\t\t\"Html\": \"%s\",\n", w.LastWork.Html)
    fmt.Fprintf(&builder, "\t\t\t\"Width\": %d,\n", w.LastWork.Width)
    fmt.Fprintf(&builder, "\t\t\t\"TimeCreate\": %s,\n", w.LastWork.Info.TimeCreate)
    fmt.Fprintf(&builder, "\t\t\t\"TimeStart\": %s,\n", w.LastWork.Info.TimeStart)
    fmt.Fprintf(&builder, "\t\t\t\"TimeEnd\": %s,\n", w.LastWork.Info.TimeEnd)

    fmt.Fprintf(&builder, "\t\t\t\"Logs\": [\n")
    for _, log := range w.LastWork.Info.Logs {
        fmt.Fprintf(&builder, "\t\t\t\t\"%s\",\n", log)
    }
    fmt.Fprintf(&builder, "\t\t\t],\n")

    fmt.Fprintf(&builder, "\t\t},\n")
    fmt.Fprintln(&builder, "\t},")
    return builder.String()
}

func (w *Worker) Printf(format string, args ...interface{}) {
    _format := fmt.Sprintf("Worker%d: %s", w.ID, format)

    log.Printf(_format, args...)
    w.LastWork.Info.Logs = append(w.LastWork.Info.Logs, fmt.Sprintf(_format, args...))
}

func (w *Worker) Start() {
	go func() {
		for {
			w.Printf("Ready to work!")
			w.WorkerQueue <- w
			select {
			case work := <-w.Work:
				// Receive a work request.
				w.Printf("Received job %s.\n", work.Name)
				var picbuf []byte

                w.LastWork = work
                w.LastWork.Info.TimeStart = time.Now().String()
				picbuf, err := w.doScreenShot(work.Ctx, w.Pt, work.Name, work.Width)
				w.Printf("Finished job %s!", work.Name)
                w.LastWork.Info.TimeEnd = time.Now().String()
				if err != nil {
					wr := WorkResponse{nil, err}
					work.Response <- wr
				} else {
					wr := WorkResponse{picbuf, nil}
					work.Response <- wr
				}

			case <-w.QuitChan:
				// We have been asked to stop.
				w.Printf("Stopping")
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.QuitChan <- true
}

var WorkerQueue chan *Worker
var WorkQueue = make(chan WorkRequest, 100)

func (p *QueuedShotter) StartDispatcher(nworkers int) {
	WorkerQueue = make(chan *Worker, nworkers)

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
		p.workers = append(p.workers, &worker)
	}

	go func() {
		for {
			work := <-WorkQueue
			log.Println("Received work requeust" + work.Name)
			go func() {
				worker := <-WorkerQueue

				log.Println("Dispatching work request")
				worker.Work <- work
			}()
		}
	}()
}

// QueuedShotter xxx
type QueuedShotter struct {
	devt         *devtool.DevTools
	workers      []*Worker
	chromeServer string
	log          *logrus.Logger
}

func (p *QueuedShotter) Stop() {
	var wg sync.WaitGroup
	for _, w := range p.workers {
		log.Println("stop worker: ", w.ID)
		wg.Add(1)
		go func(nw *Worker) {
			log.Println("stopping worker: ", nw.ID)
			p.devt.Close(context.TODO(), nw.Pt)
			wg.Done()
			log.Println("stopped worker: ", nw.ID)
		}(w)
	}
	wg.Wait()
}

// Do xxx
func (p *QueuedShotter) Do(ctx context.Context, url string, html string, width int) ([]byte, error) {
	resp := make(chan WorkResponse)
	work := NewWorkRequest(url, resp, html, width, ctx)
	log.Infof("Append job %s to chan", url)
	WorkQueue <- work
	response := <-resp
	log.Infof("Job %s finished", url)
	return response.Picbuf, response.Error
}

func (p *QueuedShotter) Stat() string {
    var builder strings.Builder
    fmt.Fprintln(&builder, "[")
    for _, w := range p.workers {
        fmt.Fprint(&builder, w.Json())
    }
    fmt.Fprintln(&builder, "]")
    return builder.String()
}

func (w *Worker) doScreenShot(ctx context.Context, pt *devtool.Target, url string, width int) ([]byte, error) {
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)
	// Open a DOMContentEventFired client to buffer this event.
	w.Printf("Open a DOMContentEventFired client to buffer this event.")
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	w.Printf("Enable events on the Page domain.")
	if err = c.Page.Enable(ctx); err != nil {
		return nil, err
	}

	// Create the Navigate arguments with the optional Referrer field set.
	w.Printf("Create the Navigate arguments with the optional Referrer field set.")
	navArgs := page.NewNavigateArgs(url)
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		return nil, err
	}

	// Wait until we have a DOMContentEventFired event.
	w.Printf("Wait until we have a DOMContentEventFired event.")
	if _, err = domContent.Recv(); err != nil {
		return nil, err
	}

	w.Printf("Page loaded with frame ID: %s", nav.FrameID)
	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	getvp := func() (*page.Viewport, error) {
		doc2, err := c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{NodeID: doc.Root.NodeID, Selector: "#ACHHcLIkD3"})
		if err != nil {
			return nil, err
		}

		c.DOM.SetAttributeValue(ctx, &dom.SetAttributeValueArgs{NodeID: doc2.NodeID, Name: "style", Value: fmt.Sprintf("width: %dpx", width)})
		rect, err2 := c.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{NodeID: &doc2.NodeID})
		if err2 != nil {
			return nil, err2
		}

		vp := page.Viewport{
			X:      rect.Model.Content[0],
			Y:      rect.Model.Content[1],
			Width:  float64(rect.Model.Width),
			Height: float64(rect.Model.Height),
			Scale:  1.0,
		}
		return &vp, nil
	}
	var vp *page.Viewport
	for step := 1; step < 3000; step *= 2 {
		vp, err = getvp()
		if err != nil {
			return nil, err
		}
		if vp.Y == 0 {
			break
		}
		w.Printf("Sleeping %dms.", step)
		time.Sleep(time.Millisecond * time.Duration(step))
	}
	w.Printf("%+v", vp)
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

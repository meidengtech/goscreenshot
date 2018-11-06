package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
)

type WorkRequest struct {
	Name     string
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
	CdpRes      *chromedp.Res
}

func NewWorker(id int, workerQueue chan chan WorkRequest, cdpRes *chromedp.Res) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		CdpRes:      cdpRes,
	}

	return worker
}

func (w *Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				fmt.Printf("worker%d: Hello, %s!\n", w.ID, work.Name)
				var picbuf []byte
				w.CdpRes.Run(context.TODO(), screenshot(work.Name, &picbuf, 750))
				wr := WorkResponse{picbuf, nil}
				work.Response <- wr

			case <-w.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	w.QuitChan <- true
	w.CdpRes.Release()
}

var WorkerQueue chan chan WorkRequest
var WorkQueue = make(chan WorkRequest, 100)

func (p *QueuedShotter) StartDispatcher(nworkers int) {
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		fmt.Println("Starting worker", i+1)
		// TODO: enable chromedp
		cdpRes, _ := p.cPool.Allocate(context.TODO(),
			runner.Flag("headless", true),
			runner.Flag("no-default-browser-check", true),
			runner.Flag("no-first-run", true),
			runner.Flag("disable-gpu", true),
			runner.Flag("no-sandbox", true))

		worker := NewWorker(i+1, WorkerQueue, cdpRes)
		worker.Start()
		p.workers = append(p.workers, worker)
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				fmt.Println("Received work requeust")
				go func() {
					worker := <-WorkerQueue

					fmt.Println("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

// QueuedShotter xxx
type QueuedShotter struct {
	cPool   *chromedp.Pool
	workNum int
	workers []Worker
}

// Start 启动dispatcher
func (p *QueuedShotter) Start() {
	p.StartDispatcher(p.workNum)
}

func (p *QueuedShotter) Stop() {
	var wg sync.WaitGroup

	for _, w := range p.workers {
		fmt.Println("stop worker: ", w.ID)
		wg.Add(1)
		go func(nw Worker) {
			nw.Stop()
			wg.Done()
			fmt.Println("stopped worker: ", nw.ID)
		}(w)
	}
	wg.Wait()
}

// Do xxx
func (p *QueuedShotter) Do(ctxt1 context.Context, url string, width int) ([]byte, error) {
	resp := make(chan WorkResponse)
	work := WorkRequest{Name: url, Response: resp}
	WorkQueue <- work
	response := <-resp
	return response.Picbuf, response.Error
}

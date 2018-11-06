package main

import (
	"context"
	"fmt"

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
	CdpPool     *chromedp.Pool
}

func NewWorker(id int, workerQueue chan chan WorkRequest, cdpPool *chromedp.Pool) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		CdpPool:     cdpPool,
	}

	return worker
}

func (w *Worker) Start() {
	go func() {
		// TODO: enable chromedp
		cdp, _ := w.CdpPool.Allocate(context.TODO(),
			runner.Flag("headless", true),
			runner.Flag("no-default-browser-check", true),
			runner.Flag("no-first-run", true),
			runner.Flag("disable-gpu", true),
			runner.Flag("no-sandbox", true))
		defer cdp.Release()
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				fmt.Printf("worker%d: Hello, %s!\n", w.ID, work.Name)
				var picbuf []byte
				cdp.Run(context.TODO(), screenshot(work.Name, &picbuf, 750))
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
	go func() {
		w.QuitChan <- true
	}()
}

var WorkerQueue chan chan WorkRequest
var WorkQueue = make(chan WorkRequest, 100)

func StartDispatcher(nworkers int, cdpPool *chromedp.Pool) {
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		fmt.Println("Starting worker", i+1)
		worker := NewWorker(i+1, WorkerQueue, cdpPool)
		worker.Start()
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

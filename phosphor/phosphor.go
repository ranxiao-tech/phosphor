package phosphor

import (
	"net/http"

	"golang.org/x/net/context"

	log "github.com/cihub/seelog"
)

type Phosphor struct {
	opts  *Options
	Store Store

	exitChan chan struct{}
}

func New(opts *Options) *Phosphor {
	return &Phosphor{
		opts: opts,
		// Store: opts.Store,

		exitChan: make(chan struct{}),
	}
}

func (p *Phosphor) Run() {
	log.Infof("Phosphor starting up")
	defer log.Flush()

	// Store a reference to phosphor in our context which we can pass
	// to other areas of the application, eg the HTTP api
	ctx := context.Background()
	ctx = context.WithValue(ctx, "phosphor", p)

	// Initialise a persistent store
	// if p.Store == nil {
	p.Store = NewMemoryStore()
	// }

	// Initialise trace ingestion
	go p.RunIngester()

	// Set up API and serve requests
	http.HandleFunc("/", Index)
	http.HandleFunc("/trace", TraceLookup(ctx))
	go http.ListenAndServe(p.opts.HTTPAddress, nil)
}

func (p *Phosphor) Exit() {
	log.Infof("Phosphor exiting")
	select {
	case <-p.exitChan: // check if already closed
	default:
		close(p.exitChan)
	}
}

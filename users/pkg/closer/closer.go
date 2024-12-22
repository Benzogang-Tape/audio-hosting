package closer

import (
	"log/slog"
	"os"
	"sync"
)

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
	log   *slog.Logger
}

func New(log *slog.Logger, sig ...os.Signal) *Closer {
	c := &Closer{
		done: make(chan struct{}),
		log:  log,
	}
	return c
}

func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

func (c *Closer) Wait() {
	<-c.done
}

func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				c.log.Error("failed to close", slog.String("error", err.Error()))
			}
		}
	})
}

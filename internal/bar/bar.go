// Package bar provides functionality for progress bar rendering.
package bar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v2"
	log "github.com/sirupsen/logrus"
)

// Type represents type of progress bar.
//go:generate stringer -type=Type -trimprefix=Type
type Type uint

const (
	typeUnknown Type = iota

	// TypeRendered is a progress bar that will be rendered.
	TypeRendered
	// TypeVoid is a void progress bar will do nothing.
	TypeVoid

	typeSentinel
)

// Valid checks if type is in a valid value range.
func (bt Type) Valid() bool {
	return bt > typeUnknown && bt < typeSentinel
}

// Bar is a progress bar manipulation contract.
type Bar interface {
	// Progress returns write channel, that will increase done work.
	Progress() chan<- struct{}
	// Finish stops the progress bar, means that no work left to do. Should be called in defer after bar created.
	Finish()
	// Run runs progress bar rendering. Blocking process, should be run in a goroutine.
	Run(ctx context.Context)
}

// New creates Bar instance for bar progress rendering.
// cap - is the expected amount of work.
// barType - is a desired type of bar that constructor will return.
// Usage:
//
// pBar := bar.New(len(notMutual), log.GetLevel())
//
// go pBar.Run(ctx)
// defer func() {
//	pBar.Finish()
// }()
//
// for i := range 100{
// 	pBar.Progress() <- struct{}{}
// }
func New(cap int, barType Type) Bar {
	switch barType {
	case TypeRendered:
		return &realBar{
			bar:   progressbar.New(cap),
			stop:  sync.Once{},
			wg:    sync.WaitGroup{},
			bchan: make(chan struct{}),
		}
	case TypeVoid:
		return &voidBar{
			wg:    sync.WaitGroup{},
			stop:  sync.Once{},
			bchan: make(chan struct{}),
		}
	default:
		return nil
	}
}

type voidBar struct {
	wg    sync.WaitGroup
	stop  sync.Once
	bchan chan struct{}
}

func (vb *voidBar) Progress() chan<- struct{} {
	return vb.bchan
}

func (vb *voidBar) Finish() {
	vb.stop.Do(func() {
		close(vb.bchan)
	})

	vb.wg.Wait()
}

func (vb *voidBar) Run(ctx context.Context) {
	vb.wg.Add(1)

	defer func() {
		vb.wg.Done()
	}()

	for {
		select {
		case _, ok := <-vb.bchan:
			if !ok {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

type realBar struct {
	bar   *progressbar.ProgressBar
	stop  sync.Once
	wg    sync.WaitGroup
	bchan chan struct{}
}

func (b *realBar) Finish() {
	b.stop.Do(func() {
		close(b.bchan)
	})

	b.wg.Wait()
}

func (b *realBar) Progress() chan<- struct{} {
	return b.bchan
}

func (b *realBar) Run(ctx context.Context) {
	b.wg.Add(1)

	defer func() {
		b.wg.Done()
	}()

	defer func() {
		if err := b.bar.Finish(); err != nil {
			log.Errorf("error when finish bar: %v", err)
		}
		fmt.Println()
	}()

	for {
		select {
		case _, ok := <-b.bchan:
			if !ok {
				return
			}

			if err := b.bar.Add(1); err != nil {
				log.Errorf("error when add to bar: %v", err)
			}

			time.Sleep(10 * time.Millisecond)
		case <-ctx.Done():
			log.Errorf("canceled context: %v", ctx.Err())
			return
		}
	}
}

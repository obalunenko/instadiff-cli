// Package bar provides functionality for progress bar rendering.
package bar

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v2"
	log "github.com/sirupsen/logrus"
)

// BType represents kind of progress bar.
//go:generate stringer -type=BType -trimprefix=BType
type BType uint

const (
	// BTypeUnknown is a default empty value for unknown progress bar type.
	BTypeUnknown BType = iota

	// BTypeRendered is a progress bar that will be rendered.
	BTypeRendered
	// BTypeVoid is a void progress bar will do nothing.
	BTypeVoid

	bTypeSentinel
)

// Valid checks if type is in a valid value range.
func (i BType) Valid() bool {
	return i > BTypeUnknown && i < bTypeSentinel
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
// max - is the expected amount of work.
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
// }.
//
func New(max int, barType BType) Bar {
	switch barType { //nolint:exhaustive
	case BTypeRendered:
		b := realBar{
			bar:   progressbar.New(max),
			stop:  sync.Once{},
			wg:    sync.WaitGroup{},
			bchan: make(chan struct{}, 1),
		}

		b.wg.Add(1)

		return &b
	case BTypeVoid:
		b := voidBar{
			wg:    sync.WaitGroup{},
			stop:  sync.Once{},
			bchan: make(chan struct{}, 1),
		}

		b.wg.Add(1)

		return &b
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
	defer func() {
		b.wg.Done()
	}()

	defer func() {
		if err := b.bar.Finish(); err != nil {
			log.Errorf("error when finish bar: %v", err)
		}

		_, _ = fmt.Fprintln(os.Stdout)
	}()

	var (
		milisecondsNum time.Duration = 10
		sleep                        = milisecondsNum * time.Millisecond
		progressInc                  = 1
	)

	for {
		select {
		case _, ok := <-b.bchan:
			if !ok {
				return
			}

			if err := b.bar.Add(progressInc); err != nil {
				log.Errorf("error when add to bar: %v", err)
			}

			time.Sleep(sleep)
		case <-ctx.Done():
			log.Errorf("canceled context: %v", ctx.Err())

			return
		}
	}
}

// Package streamdelivery records row acknowledgement for one synchronous typed
// stream invocation. It is runtime plumbing, not remotely supplied metadata.
package streamdelivery

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/teaql/teaql-golang/core"
)

var ErrStop = errors.New("stop stream")

type contextKey struct{}
type Delivery struct {
	root *core.QueryRequestScope
	rows int64
}

func New(ctx context.Context, root *core.QueryRequestScope) (context.Context, *Delivery) {
	d := &Delivery{root: root}
	return context.WithValue(ctx, contextKey{}, d), d
}

func For(ctx context.Context, root *core.QueryRequestScope) *Delivery {
	d, ok := ctx.Value(contextKey{}).(*Delivery)
	if !ok || d.root != root {
		return nil
	}
	return d
}

func (d *Delivery) Delivered() { atomic.AddInt64(&d.rows, 1) }
func (d *Delivery) Count() int { return int(atomic.LoadInt64(&d.rows)) }
func Cancelled(err error) bool {
	return errors.Is(err, ErrStop) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

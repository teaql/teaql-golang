package runtime

import (
	"context"
	"sync"
)

// A zero-value gate serializes independent graph saves without making a
// cancelled caller wait for the current owner's provider or business callback.
// The channel is allocated once per UserContext, not once per save or waiter.
type cancellableGraphSaveGate struct {
	once   sync.Once
	holder chan struct{}
}

func (g *cancellableGraphSaveGate) channel() chan struct{} {
	g.once.Do(func() { g.holder = make(chan struct{}, 1) })
	return g.holder
}

func (g *cancellableGraphSaveGate) LockContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	holder := g.channel()
	select {
	case holder <- struct{}{}:
		// Cancellation and admission can become ready together. Cancellation
		// must still win before telemetry, preparation or provider access.
		// Release even if a custom Context.Err panics during this last check.
		handedOff := false
		defer func() {
			if !handedOff {
				<-holder
			}
		}()
		if err := ctx.Err(); err != nil {
			return err
		}
		handedOff = true
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *cancellableGraphSaveGate) TryLock() bool {
	select {
	case g.channel() <- struct{}{}:
		return true
	default:
		return false
	}
}

func (g *cancellableGraphSaveGate) Unlock() {
	select {
	case <-g.channel():
	default:
		panic("unlock of an unlocked graph save gate")
	}
}

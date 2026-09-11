package capability_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sachncs/alfred/internal/capability"
)

func TestBrokerRegisterAndDispatch(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	b.Register(capability.NewFunction("hello", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`"hi"`), nil
	}))
	out, err := b.Dispatch(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(out) != `"hi"` {
		t.Fatalf("got %s", out)
	}
}

func TestBrokerDispatchUnknownReturnsErrUnknown(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	_, err := b.Dispatch(context.Background(), "missing", nil)
	if !errors.Is(err, capability.ErrUnknownCapability) {
		t.Fatalf("expected ErrUnknownCapability, got %v", err)
	}
}

func TestBrokerDispatchUnavailable(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	b.Register(capability.NewFunction("flaky", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`null`), nil
	}).WithAvailability(func(_ context.Context) bool { return false }))
	_, err := b.Dispatch(context.Background(), "flaky", nil)
	if !errors.Is(err, capability.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestBrokerUnregister(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	b.Register(capability.NewFunction("x", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	}))
	if !b.Unregister("x") {
		t.Fatalf("unregister returned false")
	}
	if b.Unregister("x") {
		t.Fatalf("second unregister returned true")
	}
}

func TestBrokerIDs(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	b.Register(capability.NewFunction("a", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	}))
	b.Register(capability.NewFunction("b", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	}))
	ids := b.IDs()
	if len(ids) != 2 {
		t.Fatalf("ids: %v", ids)
	}
	seen := map[capability.ID]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Fatalf("missing id: %v", ids)
	}
}

func TestBrokerReplace(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	var count int32
	b.Register(capability.NewFunction("counter", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		atomic.AddInt32(&count, 1)
		return json.RawMessage(`{}`), nil
	}))
	b.Register(capability.NewFunction("counter", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		atomic.AddInt32(&count, 100)
		return json.RawMessage(`{}`), nil
	}))
	if _, err := b.Dispatch(context.Background(), "counter", nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if atomic.LoadInt32(&count) != 100 {
		t.Fatalf("second handler should have run, got count=%d", count)
	}
}

func TestFunctionAvailableDefaultsToTrue(t *testing.T) {
	t.Parallel()
	f := capability.NewFunction("x", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	})
	if !f.IsAvailable(context.Background()) {
		t.Fatalf("default should be available")
	}
}

func TestFunctionID(t *testing.T) {
	t.Parallel()
	f := capability.NewFunction("foo.bar", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	})
	if f.ID() != "foo.bar" {
		t.Fatalf("id: %s", f.ID())
	}
}

func TestFunctionInvoke(t *testing.T) {
	t.Parallel()
	called := false
	f := capability.NewFunction("x", func(_ context.Context, in json.RawMessage) (json.RawMessage, error) {
		called = true
		return json.RawMessage(`{"got":"` + string(in) + `"}`), nil
	})
	out, err := f.Invoke(context.Background(), json.RawMessage(`"hi"`))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !called {
		t.Fatalf("invoke not called")
	}
	if !strings.Contains(string(out), "hi") {
		t.Fatalf("output: %s", out)
	}
}

func TestBrokerConcurrentDispatch(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	var calls int32
	b.Register(capability.NewFunction("x", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		atomic.AddInt32(&calls, 1)
		return json.RawMessage(`{}`), nil
	}))
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := b.Dispatch(context.Background(), "x", nil); err != nil {
				t.Errorf("err: %v", err)
			}
		}()
	}
	wg.Wait()
	if atomic.LoadInt32(&calls) != 100 {
		t.Fatalf("expected 100 calls, got %d", calls)
	}
}

func TestBrokerGetReturnsRegistered(t *testing.T) {
	t.Parallel()
	b := capability.NewBroker()
	f := capability.NewFunction("x", func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	})
	b.Register(f)
	if b.Get("x") != f {
		t.Fatalf("Get returned wrong capability")
	}
	if b.Get("missing") != nil {
		t.Fatalf("Get returned non-nil for missing id")
	}
}

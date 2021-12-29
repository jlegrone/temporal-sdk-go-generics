package workflow

import (
	"context"

	"github.com/jlegrone/sdk-go-generics/temporal"
	"go.temporal.io/sdk/workflow"
)

type Future[T temporal.Value] struct {
	wrapped workflow.Future
}

func (f *Future[T]) Get(ctx Context) (T, error) {
	var result T
	if err := f.wrapped.Get(ctx, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (f *Future[T]) IsReady() bool {
	return f.wrapped.IsReady()
}

// NewFuture creates a new Future as well as associated Settable that is used to set its value.
func NewFuture[T temporal.Value](ctx Context) (*Future[T], *Settable[T]) {
	fut, set := workflow.NewFuture(ctx)
	return &Future[T]{fut}, &Settable[T]{set}
}

func WrapFuture[T temporal.Value](future workflow.Future) *Future[T] {
	return &Future[T]{
		wrapped: future,
	}
}

type Settable[T temporal.Value] struct {
	wrapped workflow.Settable
}

func (s *Settable[T]) Set(value T, err error) {
	s.wrapped.Set(value, err)
}

func (s *Settable[T]) SetValue(value T) {
	s.wrapped.SetValue(value)
}

func (s *Settable[T]) SetError(err error) {
	s.wrapped.SetError(err)
}

func (s *Settable[T]) Chain(future *Future[T]) {
	s.wrapped.Chain(future.wrapped)
}

func ExecuteActivity[Resp temporal.Value](ctx Context, activity any, args ...any) *Future[Resp] {
	fut := workflow.ExecuteActivity(ctx, activity, args...)
	return WrapFuture[Resp](fut)
}

func ExecuteActivityFunc[Req, Resp temporal.Value](ctx Context, activity func(context.Context, Req) (Resp, error), req Req) *Future[Resp] {
	fut := workflow.ExecuteActivity(ctx, activity, req)
	return WrapFuture[Resp](fut)
}

func ExecuteLocalActivity[Resp temporal.Value](ctx Context, localActivity any, args ...any) *Future[Resp] {
	fut := workflow.ExecuteLocalActivity(ctx, localActivity, args...)
	return WrapFuture[Resp](fut)
}

func ExecuteLocalActivityFunc[Req, Resp temporal.Value](ctx Context, activity func(context.Context, Req) (Resp, error), req Req) *Future[Resp] {
	fut := workflow.ExecuteLocalActivity(ctx, activity, req)
	return WrapFuture[Resp](fut)
}

func ExecuteChildWorkflow[Resp temporal.Value](ctx Context, childWorkflow any, args ...any) *Future[Resp] {
	fut := workflow.ExecuteChildWorkflow(ctx, childWorkflow, args...)
	return WrapFuture[Resp](fut)
}

func ExecuteChildWorkflowFunc[Req, Resp temporal.Value](ctx Context, childWorkflow func(Context, Req) (Resp, error), req Req) *Future[Resp] {
	fut := workflow.ExecuteChildWorkflow(ctx, childWorkflow, req)
	return WrapFuture[Resp](fut)
}

type SendChannel[T temporal.Value] struct {
	wrapped workflow.SendChannel
}

// Send blocks until the data is sent.
func (sc *SendChannel[T]) Send(ctx Context, v T) {
	sc.wrapped.Send(ctx, v)
}

// SendAsync try to send without blocking. It returns true if the data was sent, otherwise it returns false.
func (sc *SendChannel[T]) SendAsync(v T) (ok bool) {
	return sc.wrapped.SendAsync(v)
}

// Close close the Channel, and prohibit subsequent sends.
func (sc *SendChannel[T]) Close() {
	sc.wrapped.Close()
}

func WrapSendChannel[T temporal.Value](sc workflow.SendChannel) *SendChannel[T] {
	return &SendChannel[T]{
		wrapped: sc,
	}
}

type ReceiveChannel[T temporal.Value] struct {
	wrapped workflow.ReceiveChannel
}

// Receive blocks until a value is sent on the channel.
// Returns false when Channel is closed.
func (rc *ReceiveChannel[T]) Receive(ctx Context) (value T, more bool) {
	more = rc.wrapped.Receive(ctx, &value)
	return
}

// ReceiveAsync tries to receive from Channel without blocking.
// Returns true when Channel has data available, otherwise it returns false immediately.
func (rc *ReceiveChannel[T]) ReceiveAsync() (value T, ok bool) {
	ok = rc.wrapped.ReceiveAsync(&value)
	return
}

// ReceiveAsyncWithMoreFlag is same as ReceiveAsync with extra return value more to indicate if there could be
// more values from the Channel. The more is false when Channel is closed.
func (rc *ReceiveChannel[T]) ReceiveAsyncWithMoreFlag() (value T, ok, more bool) {
	ok, more = rc.wrapped.ReceiveAsyncWithMoreFlag(&value)
	return
}

func WrapReceiveChannel[T temporal.Value](rc workflow.ReceiveChannel) *ReceiveChannel[T] {
	return &ReceiveChannel[T]{
		wrapped: rc,
	}
}

type Channel[T temporal.Value] struct {
	*SendChannel[T]
	*ReceiveChannel[T]
}

func WrapChannel[T temporal.Value](ch workflow.Channel) *Channel[T] {
	return &Channel[T]{
		SendChannel:    WrapSendChannel[T](ch),
		ReceiveChannel: WrapReceiveChannel[T](ch),
	}
}

// NewChannel creates new Channel instance.
func NewChannel[T temporal.Value](ctx Context) *Channel[T] {
	return WrapChannel[T](workflow.NewChannel(ctx))
}

// NewNamedChannel creates new Channel instance with a given human readable name.
// Name appears in stack traces that are blocked on this channel.
func NewNamedChannel[T temporal.Value](ctx Context, name string) *Channel[T] {
	return WrapChannel[T](workflow.NewNamedChannel(ctx, name))
}

// NewBufferedChannel creates new buffered Channel instance.
func NewBufferedChannel[T temporal.Value](ctx Context, size int) *Channel[T] {
	return WrapChannel[T](workflow.NewBufferedChannel(ctx, size))
}

// NewNamedBufferedChannel creates a new BufferedChannel instance with a given human readable name.
// Name appears in stack traces that are blocked on this Channel.
func NewNamedBufferedChannel[T temporal.Value](ctx Context, name string, size int) *Channel[T] {
	return WrapChannel[T](workflow.NewNamedBufferedChannel(ctx, name, size))
}

// SideEffect executes the provided function once, records its result into the workflow history.
//
// The recorded result on history will be returned without executing the provided function during replay.
// This guarantees the deterministic requirement for workflow as the exact same result will be returned in replay.
//
// Common use case is to run some short non-deterministic code in workflow, like getting random number or new UUID.
// The only way to fail SideEffect is to panic which causes workflow task failure. The workflow task after timeout is
// rescheduled and re-executed giving SideEffect another chance to succeed.
//
// Caution: do not use SideEffect to modify closures. Always retrieve result from SideEffect's encoded return value.
func SideEffect[T temporal.Value](ctx Context, f func(Context) T) (T, error) {
	var result T
	err := workflow.SideEffect(ctx, func(ctx Context) interface{} {
		return f(ctx)
	}).Get(&result)
	return result, err
}

// MutableSideEffect executes the provided function once, then it looks up the history for the value with the given id.
//
// If there is no existing value, then it records the function result as a value with the given id on history;
// otherwise, it compares whether the existing value from history has changed from the new function result.
// If they are equal, it returns the value without recording a new one in history; otherwise, it records the new value
// with the same id on history.
//
// Caution: do not use MutableSideEffect to modify closures. Always retrieve result from MutableSideEffect's return value.
//
// The difference between MutableSideEffect() and SideEffect() is that every new SideEffect() call in non-replay will
// result in a new marker being recorded on history. However, MutableSideEffect() only records a new marker if the value
// changed. During replay, MutableSideEffect() will not execute the function again, but it will return the exact same
// value as it was returning during the non-replay run.
//
// One good use case of MutableSideEffect() is to access dynamically changing config without breaking determinism.
func MutableSideEffect[T temporal.ComparableValue](ctx Context, id string, f func(Context) T) (T, error) {
	var result T
	err := workflow.MutableSideEffect(ctx, id,
		func(ctx Context) interface{} {
			return f(ctx)
		}, func(a, b any) bool {
			return a.(T) == b.(T)
		},
	).Get(&result)
	return result, err
}

// SelectorAddReceive is analogous to workflow.Selector.AddReceive
func selectorAddReceive[T temporal.Value](s workflow.Selector, rc *ReceiveChannel[T], callback func(val T, more bool)) workflow.Selector {
	return s.AddReceive(rc.wrapped, func(ch workflow.ReceiveChannel, more bool) {
		val, ok := WrapReceiveChannel[T](ch).ReceiveAsync()
		if !ok {
			panic("no value on channel")
		}
		callback(val, more)
	})
}

// AddReceive is analogous to workflow.Selector.AddReceive
func (rc *ReceiveChannel[T]) AddReceive(s workflow.Selector, callback func(val T, more bool)) workflow.Selector {
	return selectorAddReceive(s, rc, callback)
}

// SelectorAddSend is analogous to workflow.Selector.AddSend
func selectorAddSend[T temporal.Value](s workflow.Selector, sc *SendChannel[T], v T, f func()) workflow.Selector {
	return s.AddSend(sc.wrapped, v, f)
}

// SelectorAddSendValue sends a message without requiring a callback function.
func selectorAddSendValue[T temporal.Value](s workflow.Selector, sc *SendChannel[T], v T) workflow.Selector {
	return s.AddSend(sc.wrapped, v, func() {
		if !sc.SendAsync(v) {
			panic("message did not send")
		}
	})
}

// AddSend is analogous to workflow.Selector.AddSend
func (sc *SendChannel[T]) AddSend(s workflow.Selector, v T, f func()) workflow.Selector {
	return selectorAddSend(s, sc, v, f)
}

// AddSendValue sends a message without requiring a callback function
func (sc *SendChannel[T]) AddSendValue(s workflow.Selector, v T) workflow.Selector {
	return selectorAddSendValue(s, sc, v)
}

// SelectorAddFuture is analogous to workflow.Selector.AddFuture
func selectorAddFuture[T temporal.Value](ctx Context, s workflow.Selector, future *Future[T], callback func(T, error)) workflow.Selector {
	return s.AddFuture(future.wrapped, func(f workflow.Future) {
		val, err := WrapFuture[T](f).Get(ctx)
		callback(val, err)
	})
}

// AddFuture is analogous to workflow.Selector.AddFuture
func (f *Future[T]) AddFuture(ctx Context, s workflow.Selector, callback func(T, error)) workflow.Selector {
	return selectorAddFuture(ctx, s, f, callback)
}

func SetQueryHandler[Resp temporal.Value](ctx Context, queryType string, handler func(...temporal.Value) (Resp, error)) error {
	return workflow.SetQueryHandler(ctx, queryType, handler)
}

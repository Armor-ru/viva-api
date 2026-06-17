package viva_api

import (
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestNotify_Send(t *testing.T) {
	t.Parallel()

	tr := &fakeTransport{}
	v := &Viva{ussdTransport: tr}

	if err := v.notify("37477600552", "1020", "hello", "test"); err != nil {
		t.Fatalf("notify() error = %v", err)
	}
	if len(tr.sendCalls) != 1 {
		t.Fatalf("expected 1 send, got %d", len(tr.sendCalls))
	}
	payload, ok := tr.sendCalls[0].msg.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", tr.sendCalls[0].msg)
	}
	if payload["to"] != "37477600552" || payload["text"] != "hello" {
		t.Fatalf("unexpected payload: %v", payload)
	}
}

type sendCall struct {
	topic string
	msg   interface{}
}

type fakeTransport struct {
	sendCalls []sendCall
}

func (f *fakeTransport) Connect() error                       { return nil }
func (f *fakeTransport) ConnectAndWait() error                { return nil }
func (f *fakeTransport) Disconnect() error                    { return nil }
func (f *fakeTransport) Middleware(args ...interface{}) error { return nil }
func (f *fakeTransport) Subscribe(topic string, h types.Handler) (types.Subscriber, error) {
	return nil, nil
}
func (f *fakeTransport) Emit(topic string, msg types.Message) error { return nil }
func (f *fakeTransport) Send(topic string, msg types.Message, opt types.SendOptions) (types.Message, error) {
	f.sendCalls = append(f.sendCalls, sendCall{topic: topic, msg: msg})
	return map[string]interface{}{}, nil
}
func (f *fakeTransport) Error(h types.ErrorHandler) {}

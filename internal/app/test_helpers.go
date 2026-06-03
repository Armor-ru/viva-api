package viva_api

import "git.dev.armlab.pro/armor/sds-go/pkg/types"

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

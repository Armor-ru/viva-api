package viva_api

import (
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

type subscribeRecordingTransport struct {
	fakeTransport
	topics []string
}

func (t *subscribeRecordingTransport) Subscribe(topic string, _ types.Handler) (types.Subscriber, error) {
	t.topics = append(t.topics, topic)
	return nil, nil
}

func TestInitHandlers_SubscribesOrderCompletedOnNATS(t *testing.T) {
	t.Parallel()

	intTr := &subscribeRecordingTransport{}
	dir := t.TempDir()
	writeTestCatalog(t, dir)
	_ = New(
		WithIntTransport(intTr),
		WithCatalogDir(dir),
	)

	found := false
	for _, topic := range intTr.topics {
		if topic == "order/completed" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("subscriptions = %v", intTr.topics)
	}
}

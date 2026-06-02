package pipeline_test

import (
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
)

type fakeActions struct {
	orders []types.OrderType
	notify []string
}

func (f *fakeActions) CreateOrder(orderType types.OrderType, ctx pipeline.Context) error {
	f.orders = append(f.orders, orderType)
	return nil
}

func (f *fakeActions) Notify(ctx pipeline.Context, tplKey string) error {
	f.notify = append(f.notify, tplKey)
	return nil
}

func TestEngineMo1(t *testing.T) {
	store, err := catalog.Load("../../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	actions := &fakeActions{}
	engine := &pipeline.Engine{Catalog: store, Actions: actions}

	err = engine.Run("mo.1", pipeline.Context{
		Phone:       "37499123456",
		ProductCode: "SAFEKID",
		Lang:        "ru",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions.orders) != 1 || actions.orders[0] != types.OrderTypeNew {
		t.Fatalf("orders: %+v", actions.orders)
	}
	if len(actions.notify) != 1 || actions.notify[0] != "welcome_trial" {
		t.Fatalf("notify: %+v", actions.notify)
	}
}

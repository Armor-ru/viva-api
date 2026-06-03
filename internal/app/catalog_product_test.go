package viva_api

import "testing"

func TestProductFromCatalogByExternalID_Step12_SAFEKID(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productFromCatalogByExternalID("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" {
		t.Fatalf("externalId = %q", product.GetExternalID())
	}
	if product.GetShortNumber() != "1020" {
		t.Fatalf("shortNumber = %q", product.GetShortNumber())
	}
}

func TestProductFromCatalogByExternalID_Unknown(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	if _, err := v.productFromCatalogByExternalID("UNKNOWN"); err == nil {
		t.Fatal("expected error for unknown externalId")
	}
}

func TestProductCodeFromOrder_PrefersCustomData(t *testing.T) {
	t.Parallel()

	order := sampleCompletedOrder()
	order.CustomData["productCode"] = "SAFEKID"
	item := order.Items[0]
	item.Product.Name = "OTHER"

	if got := productCodeFromOrder(order, item); got != "SAFEKID" {
		t.Fatalf("productCode = %q", got)
	}
}

func TestCompleteOrder_Step12_UnknownExternalID(t *testing.T) {
	t.Parallel()

	order := sampleCompletedOrder()
	order.CustomData["productCode"] = "NOT_IN_CATALOG"
	v := &viva{
		catalog:         loadTestCatalog(t),
		ussdTransport: &fakeTransport{},
	}
	if err := v.completeOrder(order); err == nil {
		t.Fatal("expected error when product not in catalog")
	}
}

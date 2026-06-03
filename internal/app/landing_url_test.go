package viva_api

import (
	"strings"
	"testing"
)

func TestBuildLandingConfirmURL(t *testing.T) {
	t.Parallel()

	v := &viva{landingConfirmURL: "https://landing.test/confirm"}
	got, err := v.buildLandingConfirmURL("37477600552", "SAFEKID", "SAFEKID", "ru")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"https://landing.test/confirm",
		"phone=37477600552",
		"productName=SAFEKID",
		"productCode=SAFEKID",
		"lang=ru",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("url = %q, missing %q", got, want)
		}
	}
}

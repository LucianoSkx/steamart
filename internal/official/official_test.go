package official

import (
	"strings"
	"testing"
)

func TestListMontaURLsDoCDN(t *testing.T) {
	assets := List(440)
	if len(assets) != 4 {
		t.Fatalf("assets = %d, want 4", len(assets))
	}
	kinds := map[string]bool{}
	for _, a := range assets {
		kinds[a.Kind] = true
		if !strings.Contains(a.URL, "/steam/apps/440/") {
			t.Errorf("%s: URL = %q", a.Kind, a.URL)
		}
		if !strings.HasPrefix(a.URL, "https://shared.cloudflare.steamstatic.com/") {
			t.Errorf("%s: host inesperado %q", a.Kind, a.URL)
		}
	}
	for _, want := range []string{"grid", "hero", "logo", "capsule"} {
		if !kinds[want] {
			t.Errorf("falta o asset %q", want)
		}
	}
}

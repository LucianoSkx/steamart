package delisted

import (
	"testing"
	"time"
)

func TestParseHTML(t *testing.T) {
	htmlText := `<html><body>
<a href='https://steam-tracker.com/app/1234/'>Half-Life 2: Episode One</a>
<a href='https://steam-tracker.com/app/5678/'>Test Game™</a>
<a href='https://steam-tracker.com/app/1234/'>Half-Life 2: Episode One</a>
</body></html>`
	apps := parseHTML([]byte(htmlText))
	if len(apps) != 2 {
		t.Fatalf("parseHTML = %d apps, want 2", len(apps))
	}
	if apps[0].AppID != 1234 || apps[0].Name != "Half-Life 2: Episode One" {
		t.Errorf("apps[0] = %+v", apps[0])
	}
	if apps[1].AppID != 5678 || apps[1].Name != "Test Game" {
		t.Errorf("apps[1] = %+v", apps[1])
	}
}

func TestResolveAppID(t *testing.T) {
	apps := []App{
		{AppID: 10, Name: "Crysis"},
		{AppID: 20, Name: "Crysis Demo"},
		{AppID: 30, Name: "A Different Game"},
	}
	if got := ResolveAppID("Crysis", apps); got != 10 {
		t.Errorf("ResolveAppID(Crysis) = %d, want 10", got)
	}
	if got := ResolveAppID("Crysis™", apps); got != 10 {
		t.Errorf("ResolveAppID(Crysis™) = %d, want 10", got)
	}
	if got := ResolveAppID("Nada Parecido", apps); got != 0 {
		t.Errorf("ResolveAppID(Nada Parecido) = %d, want 0", got)
	}
}

func TestFresh(t *testing.T) {
	idx := &Index{FetchedAt: time.Now(), Apps: []App{{AppID: 1, Name: "X"}}}
	if !idx.Fresh() {
		t.Error("índice recém-baixado deve ser fresco")
	}
	old := &Index{FetchedAt: time.Now().Add(-8 * 24 * time.Hour), Apps: []App{{AppID: 1, Name: "X"}}}
	if old.Fresh() {
		t.Error("índice com 8 dias não deve ser fresco")
	}
}

package title

import "testing"

func TestNormaliseTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Grand Theft Auto V", "grand theft auto v"},
		{"GTA: V - Enhanced", "gta v enhanced"},
		{"The Elder Scrolls V: Skyrim Special Edition", "elder scrolls v skyrim special"},
		{"Crysis™ Remastered (EU)", "crysis"},
		{"Half-Life 2: Episode One", "half life 2 episode one"},
		{"Portal 2", "portal 2"},
		{"Rock & Roll", "rock roll"},
	}
	for _, c := range cases {
		if got := NormaliseTitle(c.in); got != c.want {
			t.Errorf("NormaliseTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsNonPrimaryTitle(t *testing.T) {
	nonPrimary := []string{"Crysis Demo", "Half-Life 2: Episode One OST", "Skyrim DLC Pack", "Portal 2 Beta", "GTA V Soundtrack"}
	for _, n := range nonPrimary {
		if !IsNonPrimaryTitle(n) {
			t.Errorf("IsNonPrimaryTitle(%q) = false, want true", n)
		}
	}
	primary := []string{"Crysis", "Half-Life 2", "Portal 2", "Grand Theft Auto V"}
	for _, n := range primary {
		if IsNonPrimaryTitle(n) {
			t.Errorf("IsNonPrimaryTitle(%q) = true, want false", n)
		}
	}
}

func TestDistinctiveTokensPresent(t *testing.T) {
	if !DistinctiveTokensPresent("half life 2 episode one", "half life 2 episode one") {
		t.Error("mesma string deve casar")
	}
	if DistinctiveTokensPresent("portal 2", "portal") {
		t.Error("portal 2 não deve casar com portal")
	}
}

func TestLevRatio(t *testing.T) {
	if LevRatio("crysis", "crysis") != 1 {
		t.Error("idênticas devem dar 1")
	}
	if LevRatio("crysis", "farcry") > 0.6 {
		t.Error("strings diferentes devem dar ratio baixo")
	}
}

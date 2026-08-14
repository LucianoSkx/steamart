package match

import "testing"

func TestPickBest(t *testing.T) {
	cases := []struct {
		term    string
		results []SearchResult
		want    int
	}{
		{"Crysis", []SearchResult{{AppID: 1, Name: "Crysis Demo"}, {AppID: 2, Name: "Crysis"}}, 2},
		{"Skyrim", []SearchResult{{AppID: 3, Name: "The Elder Scrolls V: Skyrim Special Edition"}}, 3},
		{"GTA V", []SearchResult{{AppID: 4, Name: "Grand Theft Auto V"}, {AppID: 5, Name: "Grand Theft Auto V - Soundtrack"}}, 4},
		{"GTA V", []SearchResult{{AppID: 6, Name: "Grand Theft Auto V"}, {AppID: 7, Name: "Grand Theft Auto V: Official Score"}}, 6},
		{"Baldur's Gate 3", []SearchResult{{AppID: 8, Name: "Baldurs Gate 3"}}, 8},
		{"Stardew Valley", []SearchResult{{AppID: 9, Name: "Crysis"}}, 0},
	}
	for _, c := range cases {
		got := pickBest(c.term, c.results)
		if c.want == 0 && got != nil {
			t.Errorf("pickBest(%q) = %+v, want nil", c.term, got)
			continue
		}
		if c.want != 0 && (got == nil || got.AppID != c.want) {
			t.Errorf("pickBest(%q) = %+v, want appid %d", c.term, got, c.want)
		}
	}
}

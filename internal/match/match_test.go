package match

import "testing"

func TestNorm(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Grand Theft Auto V", "grand theft auto v"},
		{"GTA: V - Enhanced", "gta v enhanced"},
		{"Rock & Roll", "rock and roll"},
		{"  Spaces  Here ", "spaces here"},
	}
	for _, c := range cases {
		if got := norm(c.in); got != c.want {
			t.Errorf("norm(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestScore(t *testing.T) {
	if s := score("GTA V", "Grand Theft Auto V"); s < 0.2 {
		t.Errorf("score GTA V = %v, want >= 0.2", s)
	}
	if s := score("Baldur's Gate 3", "Baldurs Gate 3"); s != 1 {
		t.Errorf("score identical = %v, want 1", s)
	}
	if s := score("Cyberpunk 2077", "Stardew Valley"); s > 0.2 {
		t.Errorf("score unrelated = %v, want ~0", s)
	}
}

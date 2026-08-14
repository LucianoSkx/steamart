package artwork

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuffix(t *testing.T) {
	cases := map[string]string{
		"grid":   "p",
		"hero":   "_hero",
		"logo":   "_logo",
		"icon":   "_icon",
		"capsule": "",
		"":       "p",
	}
	for in, want := range cases {
		if got := Suffix(in); got != want {
			t.Errorf("Suffix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtFromURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://x/a/cover.png", ".png"},
		{"https://x/a/b.jpg", ".jpg"},
		{"https://x/a/b.jpeg", ".jpg"},
		{"https://x/a/b.webp", ".webp"},
		{"https://x/a/b.gif", ".gif"},
		{"https://x/a/b", ".jpg"},
	}
	for _, c := range cases {
		if got := extFromURL(c.in); got != c.want {
			t.Errorf("extFromURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBackupExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "123p.png")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	backupExisting(dir, "123p.png")
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("arquivo original não foi movido: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "backup", "123p.png")); err != nil {
		t.Fatalf("backup não encontrado: %v", err)
	}
}

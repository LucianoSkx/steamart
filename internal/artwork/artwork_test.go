package artwork

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuffix(t *testing.T) {
	cases := map[string]string{
		"grid":    "p",
		"hero":    "_hero",
		"logo":    "_logo",
		"icon":    "_icon",
		"capsule": "",
		"":        "p",
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

func TestInstallFazBackupEEscreveAtomico(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "123p.png")
	if err := os.WriteFile(src, []byte("antiga"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := install(dir, "123p.png", []byte("nova")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("nova arte não gravada: %v", err)
	}
	if string(got) != "nova" {
		t.Errorf("conteúdo = %q, want %q", got, "nova")
	}
	if _, err := os.Stat(filepath.Join(dir, "backup", "123p.png")); err != nil {
		t.Fatalf("backup não encontrado: %v", err)
	}
	// segunda gravação não pode descartar o primeiro backup
	if err := install(dir, "123p.png", []byte("nova2")); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "backup", "123p.png")); err != nil || string(b) != "antiga" {
		t.Errorf("primeiro backup destruído: %q, err=%v", b, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "backup", "123p-2.png")); err != nil {
		t.Errorf("segundo backup ausente: %v", err)
	}
}

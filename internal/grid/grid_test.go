package grid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBelongsToShortcut(t *testing.T) {
	cases := []struct {
		name  string
		appid uint32
		want  bool
	}{
		{"123p.png", 123, true},
		{"123.png", 123, true},
		{"123_hero.jpg", 123, true},
		{"123_logo.png", 123, true},
		{"123_icon.png", 123, true},
		{"123.webp", 123, true},
		{"1234p.png", 123, false}, // colisão de prefixo
		{"1234.png", 123, false},  // idem
		{"123.txt", 123, false},   // extensão não-imagem
		{"123", 123, false},       // sem extensão
		{"123p.png", 1234, false}, // atalho maior não casa com menor
		{"999p.png", 123, false},  // outro atalho
		{"backup", 123, false},    // diretório sem extensão
	}
	for _, c := range cases {
		if got := BelongsToShortcut(c.name, c.appid); got != c.want {
			t.Errorf("BelongsToShortcut(%q, %d) = %v, want %v", c.name, c.appid, got, c.want)
		}
	}
}

func TestMoveToBackupPreservaAnteriores(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "10p.png"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MoveToBackup(dir, "10p.png"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "10p.png"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MoveToBackup(dir, "10p.png"); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "backup", "10p.png")); err != nil || string(b) != "v1" {
		t.Errorf("primeiro backup = %q, err=%v; want v1", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "backup", "10p-2.png")); err != nil || string(b) != "v2" {
		t.Errorf("segundo backup = %q, err=%v; want v2", b, err)
	}
}

func TestRemoveSoTocaNoAtalho(t *testing.T) {
	dir := t.TempDir()
	files := map[string]bool{
		"123p.png":     true,
		"123.png":      true,
		"123_hero.jpg": true,
		"1234p.png":    false, // outro atalho (prefixo)
		"999p.png":     false,
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := Remove(dir, 123)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 3 {
		t.Errorf("removed = %d, want 3", removed)
	}
	for name, shouldGo := range files {
		_, err := os.Stat(filepath.Join(dir, name))
		if shouldGo && !os.IsNotExist(err) {
			t.Errorf("%s deveria ter sido movido", name)
		}
		if !shouldGo && os.IsNotExist(err) {
			t.Errorf("%s não deveria ter sido tocado", name)
		}
	}
}

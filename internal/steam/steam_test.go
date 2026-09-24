package steam

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func writeShortcuts(t *testing.T, dir string) {
	t.Helper()
	var b []byte
	putMap := func(key string) { b = append(b, 0x00); b = append(b, key...); b = append(b, 0) }
	putStr := func(key, val string) {
		b = append(b, 0x01)
		b = append(b, key...)
		b = append(b, 0)
		b = append(b, val...)
		b = append(b, 0)
	}
	putInt := func(key string, v int32) {
		b = append(b, 0x02)
		b = append(b, key...)
		b = append(b, 0)
		var n [4]byte
		binary.LittleEndian.PutUint32(n[:], uint32(v))
		b = append(b, n[:]...)
	}
	putMap("shortcuts")
	putMap("0")
	putInt("appid", 313373)
	putStr("AppName", "DOSBox")
	putStr("Exe", `"/usr/bin/dosbox"`)
	putStr("StartDir", `"/home/user"`)
	putMap("tags")
	putStr("0", "emulador")
	b = append(b, 0x08) // fim tags
	b = append(b, 0x08) // fim atalho
	b = append(b, 0x08) // fim shortcuts
	b = append(b, 0x08) // fim raiz
	if err := os.WriteFile(filepath.Join(dir, "shortcuts.vdf"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestShortcuts(t *testing.T) {
	dir := t.TempDir()
	writeShortcuts(t, dir)
	s := &Steam{Config: dir, Grid: filepath.Join(dir, "grid")}
	list, err := s.Shortcuts()
	if err != nil {
		t.Fatalf("Shortcuts: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("atalhos = %d, want 1", len(list))
	}
	sc := list[0]
	if sc.AppID != 313373 || sc.AppName != "DOSBox" || sc.Exe != `"/usr/bin/dosbox"` {
		t.Errorf("atalho = %+v", sc)
	}
	if len(sc.Tags) != 1 || sc.Tags[0] != "emulador" {
		t.Errorf("tags = %v", sc.Tags)
	}
}

func TestShortcutsVDFInvalido(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "shortcuts.vdf"), []byte("lixo"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := &Steam{Config: dir}
	if _, err := s.Shortcuts(); err == nil {
		t.Error("VDF inválido deveria dar erro")
	}
}

func TestGridPathENaoConfundeAppIDs(t *testing.T) {
	dir := t.TempDir()
	grid := filepath.Join(dir, "grid")
	if err := os.MkdirAll(grid, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"123p.png", "1234p.png", "123.jpg"} {
		if err := os.WriteFile(filepath.Join(grid, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s := &Steam{Grid: grid}
	if got := s.GridPath(123, "p"); got != filepath.Join(grid, "123p.png") {
		t.Errorf("GridPath(123,p) = %q", got)
	}
	if got := s.GridPath(1234, "p"); got != filepath.Join(grid, "1234p.png") {
		t.Errorf("GridPath(1234,p) = %q", got)
	}
	if got := s.GridPath(123, ""); got != filepath.Join(grid, "123.jpg") {
		t.Errorf("GridPath(123,) = %q", got)
	}
	if got := s.GridPath(999, "p"); got != "" {
		t.Errorf("GridPath(999,p) = %q, want vazio", got)
	}
	art := s.HasArtwork(123)
	if art["p"] == "" || art[""] == "" || art["_hero"] != "" {
		t.Errorf("HasArtwork(123) = %v", art)
	}
}

func TestSaveJSONPermissao0600(t *testing.T) {
	dir := t.TempDir()
	s := &Steam{Config: dir}
	if err := s.SaveJSON("steamart-sgdb.json", map[string]string{"key": "abc"}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(dir, "steamart-sgdb.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("permissão = %v, want 0600", fi.Mode().Perm())
	}
	var out map[string]string
	if err := s.LoadJSON("steamart-sgdb.json", &out); err != nil {
		t.Fatal(err)
	}
	if out["key"] != "abc" {
		t.Errorf("out = %v", out)
	}
}

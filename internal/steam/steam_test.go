package steam

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func fakeProfile(t *testing.T, root, profile string, mtime time.Time) {
	t.Helper()
	cfg := filepath.Join(root, "userdata", profile, "config")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(cfg, "shortcuts.vdf")
	if err := os.WriteFile(p, []byte{0x08}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestSteamRootsIncluiSnapEEtual(t *testing.T) {
	roots := steamRoots("/steam-root", "/home/u")
	want := []string{
		"/steam-root",
		"/home/u/.steam/steam",
		"/home/u/.local/share/Steam",
		"/home/u/.var/app/com.valvesoftware.Steam/.local/share/Steam",
		"/home/u/snap/steam/common/.local/share/Steam",
		"/home/u/snap/steam/current/.local/share/Steam",
	}
	if len(roots) != len(want) {
		t.Fatalf("roots = %v", roots)
	}
	for i := range want {
		if roots[i] != want[i] {
			t.Errorf("roots[%d] = %q, want %q", i, roots[i], want[i])
		}
	}
}

func TestDiscoverEscolhePerfilMaisRecente(t *testing.T) {
	home := t.TempDir()
	antigo := time.Now().Add(-48 * time.Hour)
	recente := time.Now().Add(-time.Hour)
	// perfil antigo no root nativo, recente no Snap
	fakeProfile(t, filepath.Join(home, ".local", "share", "Steam"), "111", antigo)
	fakeProfile(t, filepath.Join(home, "snap", "steam", "common", ".local", "share", "Steam"), "222", recente)

	s, err := discover("", home)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if s.UserID != "222" {
		t.Errorf("UserID = %q, want 222 (perfil mais recente)", s.UserID)
	}
	if !strings.Contains(s.Config, filepath.Join("snap", "steam", "common")) {
		t.Errorf("Config = %q, want caminho do snap", s.Config)
	}
}

func TestDiscoverSTEAMROOTComPrecedencia(t *testing.T) {
	home := t.TempDir()
	recente := time.Now()
	fakeProfile(t, filepath.Join(home, ".local", "share", "Steam"), "111", time.Now().Add(-time.Hour))
	root := t.TempDir()
	fakeProfile(t, root, "999", recente)

	s, err := discover(root, home)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if s.UserID != "999" || s.Root != root {
		t.Errorf("s = %+v, want perfil do STEAM_ROOT", s)
	}
}

func TestDiscoverSemSteam(t *testing.T) {
	if _, err := discover("", t.TempDir()); err == nil {
		t.Error("sem Steam deveria dar erro")
	}
}

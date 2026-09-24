package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenInexistente(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(s.Matches) != 0 {
		t.Errorf("matches = %d, want 0", len(s.Matches))
	}
	if s.Recovered != "" {
		t.Errorf("Recovered = %q, want vazio", s.Recovered)
	}
}

func TestSaveOpenRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	s.Set(&Match{ShortcutAppID: 42, SteamAppID: 440, Name: "Jogo"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := again.Get(42)
	if m == nil || m.SteamAppID != 440 {
		t.Fatalf("match não recuperado: %+v", m)
	}
	// não pode sobrar temporário de escrita atômica
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("arquivo temporário órfão: %s", e.Name())
		}
	}
}

func TestOpenCorrompidoPreservaArquivo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, file)
	if err := os.WriteFile(path, []byte("{isso não é json"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open não deveria falhar em JSON inválido: %v", err)
	}
	if s.Recovered == "" {
		t.Error("Recovered deveria ser preenchido")
	}
	if len(s.Matches) != 0 {
		t.Errorf("matches = %d, want 0", len(s.Matches))
	}
	// o original não pode mais existir no caminho (foi preservado)
	if _, err := os.Stat(path); err == nil {
		t.Error("arquivo corrompido ainda no lugar: um Save o destruiria")
	}
	corrompidos, err := filepath.Glob(path + ".corrompido-*")
	if err != nil || len(corrompidos) != 1 {
		t.Errorf("cópia preservada não encontrada: %v (err=%v)", corrompidos, err)
	}
	// e a partir daí um Save não pode mais corromper nada
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if again.Recovered != "" {
		t.Errorf("store recém-salvo não deveria estar corrompido: %s", again.Recovered)
	}
}

func TestOpenArquivoVazio(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, file), []byte("  \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(s.Matches) != 0 {
		t.Errorf("matches = %d, want 0", len(s.Matches))
	}
}

package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"steamart/internal/atomicfile"
)

const file = "steamart-matches.json"

// Match registra a associação de um atalho a um app Steam.
type Match struct {
	ShortcutAppID uint32 `json:"shortcut_appid"`
	SteamAppID    int    `json:"steam_appid"`
	Name          string `json:"name"`
	Pinned        bool   `json:"pinned"`
}

// Store mantém as associações em memória e em disco.
type Store struct {
	mu      sync.Mutex
	path    string
	Matches map[uint32]*Match `json:"matches"`

	// Recovered descreve uma recuperação feita na abertura (ex.: arquivo
	// corrompido preservado como .corrompido). Vazio quando nada ocorreu.
	Recovered string
}

// Open carrega os matches de disco. Arquivo inexistente abre vazio. Arquivo
// com JSON inválido NÃO é sobrescrito: ele é preservado em
// <arquivo>.corrompido e o store abre vazio, com Store.Recovered preenchido.
func Open(configDir string) (*Store, error) {
	s := &Store{
		path:    filepath.Join(configDir, file),
		Matches: map[uint32]*Match{},
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return s, nil
	}
	var loaded map[uint32]*Match
	if err := json.Unmarshal(b, &loaded); err != nil {
		broken := fmt.Sprintf("%s.corrompido-%s", s.path, time.Now().Format("20060102-150405"))
		if rerr := os.Rename(s.path, broken); rerr != nil {
			return nil, fmt.Errorf("store corrompido (%s) e não foi possível preservá-lo: %w", s.path, err)
		}
		s.Recovered = fmt.Sprintf("%s inválido (%v); cópia preservada em %s e matches iniciados vazios", file, err, broken)
		return s, nil
	}
	if loaded != nil {
		s.Matches = loaded
	}
	return s, nil
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.MarshalIndent(s.Matches, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(s.path, b, 0o644)
}

func (s *Store) Set(m *Match) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Matches[m.ShortcutAppID] = m
}

func (s *Store) Get(appid uint32) *Match {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Matches[appid]
}

// Logger registra mensagens em memória (e opcionalmente em arquivo) para
// exibição na UI e depuração.
type Logger struct {
	mu  sync.Mutex
	buf []string
	f   *os.File
}

// SetFile passa um arquivo onde os logs também são gravados.
func (l *Logger) SetFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	l.f = f
	return nil
}

func (l *Logger) Add(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	line := time.Now().Format("2006-01-02 15:04:05") + " " + msg
	l.buf = append(l.buf, line)
	if len(l.buf) > 500 {
		l.buf = l.buf[len(l.buf)-500:]
	}
	if l.f != nil {
		_, _ = fmt.Fprintln(l.f, line)
	}
}

func (l *Logger) All() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.buf))
	copy(out, l.buf)
	return out
}

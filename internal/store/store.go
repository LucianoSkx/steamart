package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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
}

func Open(configDir string) (*Store, error) {
	s := &Store{
		path:    filepath.Join(configDir, file),
		Matches: map[uint32]*Match{},
	}
	b, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(b, &s.Matches)
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
	return os.WriteFile(s.path, b, 0o644)
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

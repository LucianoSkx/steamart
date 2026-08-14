package steam

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"steamart/internal/vdf"
)

// Shortcut representa um atalho não-Steam do Steam.
type Shortcut struct {
	AppID    uint32   `json:"appid"`
	AppName  string   `json:"appname"`
	Exe      string   `json:"exe"`
	StartDir string   `json:"startdir"`
	Icon     string   `json:"icon"`
	Tags     []string `json:"tags"`
}

// Steam agrupa descoberta de caminhos e leitura de atalhos.
type Steam struct {
	Root     string
	UserID   string
	UserData string
	Config   string
	Grid     string
}

// Discover localiza a instalação da Steam no CachyOS/desktop Linux.
func Discover() (*Steam, error) {
	roots := []string{}
	if v := os.Getenv("STEAM_ROOT"); v != "" {
		roots = append(roots, v)
	}
	if h, err := os.UserHomeDir(); err == nil {
		roots = append(roots,
			filepath.Join(h, ".steam", "steam"),
			filepath.Join(h, ".local", "share", "Steam"),
			// Steam instalado via Flatpak (comum em distros como Fedora, Pop!_OS, Endeavour)
			filepath.Join(h, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
		)
	}

	for _, r := range roots {
		root, err := filepath.EvalSymlinks(r)
		if err != nil {
			root = r
		}
		ud := filepath.Join(root, "userdata")
		entries, err := os.ReadDir(ud)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			cfg := filepath.Join(ud, e.Name(), "config")
			if _, err := os.Stat(filepath.Join(cfg, "shortcuts.vdf")); err != nil {
				continue
			}
			return &Steam{
				Root:     root,
				UserID:   e.Name(),
				UserData: ud,
				Config:   cfg,
				Grid:     filepath.Join(cfg, "grid"),
			}, nil
		}
	}
	return nil, fmt.Errorf("instalação da Steam com atalhos não encontrada")
}

// Shortcuts lê e devolve os atalhos não-Steam.
func (s *Steam) Shortcuts() ([]Shortcut, error) {
	data, err := os.ReadFile(filepath.Join(s.Config, "shortcuts.vdf"))
	if err != nil {
		return nil, err
	}
	root, err := vdf.Parse(data)
	if err != nil {
		return nil, err
	}
	sc := root.Dict("shortcuts")
	if sc == nil {
		return nil, nil
	}
	var out []Shortcut
	for _, c := range sc.Children {
		d, ok := c.Value.(*vdf.Node)
		if !ok {
			continue
		}
		var tags []string
		if td := d.Dict("tags"); td != nil {
			for _, tc := range td.Children {
				if s, ok := tc.Value.(string); ok {
					tags = append(tags, s)
				}
			}
		}
		out = append(out, Shortcut{
			AppID:    uint32(d.Int("appid")),
			AppName:  d.Str("AppName"),
			Exe:      d.Str("Exe"),
			StartDir: d.Str("StartDir"),
			Icon:     d.Str("icon"),
			Tags:     tags,
		})
	}
	return out, nil
}

// GridPath devolve o caminho local do arquivo de arte na grid para um
// sufixo (p, _hero, _logo, _icon ou "" para a capa horizontal), ou string
// vazia se não existir (testa .png e .jpg).
func (s *Steam) GridPath(appid uint32, suffix string) string {
	for _, ext := range []string{".png", ".jpg"} {
		name := fmt.Sprintf("%d%s%s", appid, suffix, ext)
		if s.fileExists(name) {
			return filepath.Join(s.Grid, name)
		}
	}
	return ""
}

// HasArtwork retorna, para cada tipo de arte, a URL local (/grid/...) se
// o arquivo já existir na pasta grid, ou string vazia caso contrário.
func (s *Steam) HasArtwork(appid uint32) map[string]string {
	out := map[string]string{}
	for _, n := range []string{"p", "_hero", "_logo", "_icon", ""} {
		if p := s.GridPath(appid, n); p != "" {
			out[n] = "/grid/" + filepath.Base(p)
		}
	}
	return out
}

func (s *Steam) fileExists(name string) bool {
	_, err := os.Stat(filepath.Join(s.Grid, name))
	return err == nil
}

// SaveJSON grava um arquivo JSON dentro do diretório de config do usuário.
// Permissão 0600 porque pode conter a chave da API do SteamGridDB.
func (s *Steam) SaveJSON(name string, v any) error {
	if err := os.MkdirAll(s.Config, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.Config, name), b, 0o600)
}

// LoadJSON lê um arquivo JSON do diretório de config do usuário.
func (s *Steam) LoadJSON(name string, v any) error {
	b, err := os.ReadFile(filepath.Join(s.Config, name))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

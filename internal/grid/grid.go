// Pacote grid centraliza as regras de arquivos na pasta grid do Steam:
// quem pertence a qual atalho, backup sem perda e serialização de escrita.
package grid

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// mu serializa backup + gravação na grid: downloads e remoções podem rodar
// em goroutines concorrentes para o mesmo atalho.
var mu sync.Mutex

// Lock/Unlock expõem a seção crítica da grid para escritores externos.
func Lock()   { mu.Lock() }
func Unlock() { mu.Unlock() }

// BelongsToShortcut indica se um arquivo da grid pertence ao atalho, evitando
// colisão de prefixo (123 não pode casar com 1234p.png): confere o 1º byte
// após o appid e a extensão de imagem.
func BelongsToShortcut(name string, appid uint32) bool {
	prefix := strconv.FormatUint(uint64(appid), 10)
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	rest := name[len(prefix):]
	if rest == "" {
		return false
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
	default:
		return false
	}
	return rest[0] == '.' || rest[0] == 'p' || rest[0] == '_'
}

// MoveToBackup move um arquivo da grid para grid/backup/ preservando a arte
// anterior já salva lá (sufixo numérico em vez de sobrescrever). Usa o lock
// da grid; para uma operação composta (backup + escrita), use Lock/Unlock e
// BackupLocked.
func MoveToBackup(gridDir, name string) error {
	mu.Lock()
	defer mu.Unlock()
	return moveToBackup(gridDir, name)
}

// BackupLocked move um arquivo para grid/backup/. O chamador deve segurar
// Lock — é o caminho para operações compostas sem janela de corrida.
func BackupLocked(gridDir, name string) error {
	return moveToBackup(gridDir, name)
}

func moveToBackup(gridDir, name string) error {
	src := filepath.Join(gridDir, name)
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	bk := filepath.Join(gridDir, "backup")
	if err := os.MkdirAll(bk, 0o755); err != nil {
		return err
	}
	if err := os.Rename(src, filepath.Join(bk, uniqueName(bk, name))); err != nil {
		return fmt.Errorf("não consegui fazer backup de %s: %w", name, err)
	}
	return nil
}

// Remove move para grid/backup/ todos os arquivos do atalho e devolve quantos
// foram movidos.
func Remove(gridDir string, appid uint32) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	entries, err := os.ReadDir(gridDir)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() || !BelongsToShortcut(e.Name(), appid) {
			continue
		}
		if err := moveToBackup(gridDir, e.Name()); err == nil {
			removed++
		}
	}
	return removed, nil
}

// uniqueName devolve um nome livre em dir para name, acrescentando -2, -3...
// antes da extensão quando o nome já estiver em uso.
func uniqueName(dir, name string) string {
	if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 2; i <= 999; i++ {
		cand := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return filepath.Base(cand)
		}
	}
	return fmt.Sprintf("%s-%d%s", base, time.Now().UnixNano(), ext)
}

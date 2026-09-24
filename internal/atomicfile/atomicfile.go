// Pacote atomicfile grava arquivos de forma atômica: conteúdo em arquivo
// temporário no mesmo diretório e rename sobre o destino, para nunca deixar
// o destino truncado ou parcial (importante para a Steam que lê a grid ao
// vivo e para arquivos de estado do app).
package atomicfile

import (
	"os"
	"path/filepath"
)

// WriteFile grava data em path atomicamente (tmp + rename), criando o
// diretório de destino se necessário.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil {
		_ = os.Remove(tmpName)
		return werr
	}
	if cerr != nil {
		_ = os.Remove(tmpName)
		return cerr
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

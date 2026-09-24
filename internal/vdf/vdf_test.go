package vdf

import (
	"encoding/binary"
	"testing"
)

// buildShortcuts monta um shortcuts.vdf sintético no formato binário da Steam.
func buildShortcuts() []byte {
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
	putInt("appid", 123456789)
	putStr("AppName", "Meu Jogo")
	putStr("Exe", `"/usr/bin/jogo"`)
	putInt("IsHidden", 0)
	b = append(b, 0x08) // fim do atalho
	b = append(b, 0x08) // fim de shortcuts
	b = append(b, 0x08) // fim da raiz
	return b
}

func TestParseShortcuts(t *testing.T) {
	root, err := Parse(buildShortcuts())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sc := root.Dict("shortcuts")
	if sc == nil {
		t.Fatal("shortcuts ausente")
	}
	first := sc.Dict("0")
	if first == nil {
		t.Fatal("atalho 0 ausente")
	}
	if got := first.Int("appid"); got != 123456789 {
		t.Errorf("appid = %d, want 123456789", got)
	}
	if got := first.Str("AppName"); got != "Meu Jogo" {
		t.Errorf("AppName = %q, want %q", got, "Meu Jogo")
	}
	if got := first.Str("Exe"); got != `"/usr/bin/jogo"` {
		t.Errorf("Exe = %q", got)
	}
}

func TestMarshalRoundtrip(t *testing.T) {
	root, err := Parse(buildShortcuts())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	again, err := Parse(Marshal(root))
	if err != nil {
		t.Fatalf("Parse(Marshal): %v", err)
	}
	if len(again.Children) != len(root.Children) {
		t.Fatalf("children = %d, want %d", len(again.Children), len(root.Children))
	}
	a := again.Dict("shortcuts").Dict("0")
	b := root.Dict("shortcuts").Dict("0")
	if a.Str("AppName") != b.Str("AppName") || a.Int("appid") != b.Int("appid") {
		t.Errorf("roundtrip divergiu: %+v vs %+v", a, b)
	}
}

func TestParseTruncado(t *testing.T) {
	full := buildShortcuts()
	if _, err := Parse(full[:len(full)-1]); err == nil {
		t.Error("arquivo truncado deveria dar erro")
	}
	// int32 cortado (falta 1 byte do appid, que começa em 21)
	cut := buildShortcuts()
	if _, err := Parse(cut[:24]); err == nil {
		t.Error("int32 truncado deveria dar erro")
	}
}

func TestParseVazio(t *testing.T) {
	if _, err := Parse(nil); err == nil {
		t.Error("entrada vazia deveria dar erro")
	}
}

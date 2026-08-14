// Pacote i18n fornece tradução simples pt-BR/en para a interface.
package i18n

import (
	"fmt"
)

// Lang identifica um idioma suportado.
type Lang string

const (
	PTBR Lang = "pt-BR"
	EN   Lang = "en"
)

var current = PTBR

// Set define o idioma ativo.
func Set(l Lang) { current = l }

// Get retorna o idioma ativo.
func Get() Lang { return current }

type entry struct {
	pt string
	en string
}

var dict = map[string]entry{
	"key_placeholder":    {"Cole aqui a SteamGridDB API key…", "Paste your SteamGridDB API key here…"},
	"save_key":           {"Salvar", "Save"},
	"warning":            {"Aviso", "Warning"},
	"key_empty":          {"Informe a key antes de salvar.", "Enter the key before saving."},
	"ok":                 {"OK", "OK"},
	"key_saved":          {"SteamGridDB key salva.", "SteamGridDB key saved."},
	"refresh":            {"Atualizar", "Refresh"},
	"key_label":          {"API Key:", "API Key:"},
	"status":             {"Steam: %s · %d atalho(s) · %d pendente(s)", "Steam: %s · %d shortcut(s) · %d pending"},
	"no_shortcuts":       {"Nenhum atalho não-Steam encontrado.", "No non-Steam shortcut found."},
	"meta_none":          {"Sem metadados Steam aplicados", "No Steam metadata applied"},
	"meta":               {"Metadados: %s (app %d)", "Metadata: %s (app %d)"},
	"no_art":             {"sem arte na grid", "no art in grid"},
	"open_sgdb":          {"Buscar imagens", "Search images"},
	"auto_steam":         {"Auto Steam", "Auto Steam"},
	"details":            {"Detalhes", "Details"},
	"details_title":      {"Detalhes de %s", "Details for %s"},
	"need_match_details": {"Casa o jogo primeiro (Auto Steam ou Buscar imagens).", "Match the game first (Auto Steam or Search images)."},
	"no_store_page":      {"Sem página na loja (delisted ou não listado).", "No store page (delisted or unlisted)."},
	"meta_desc":          {"Descrição", "Description"},
	"meta_devs":          {"Desenvolvedor(es)", "Developer(s)"},
	"meta_pubs":          {"Publicador(es)", "Publisher(s)"},
	"meta_release":       {"Lançamento", "Release"},
	"meta_genres":        {"Gêneros", "Genres"},
	"meta_cats":          {"Categorias", "Categories"},
	"meta_deck":          {"Steam Deck", "Steam Deck"},
	"meta_website":       {"Site", "Website"},
	"meta_screens":       {"Screenshots", "Screenshots"},
	"remove_art":         {"Limpar imagens", "Clear images"},
	"cdn_unavailable":    {"Nenhum asset baixado (CDN indisponível?).", "No asset downloaded (CDN unavailable?)."},
	"log_art":            {"arte aplicada a %s (app %d)", "art applied to %s (app %d)"},
	"auto_no_match":      {"Não encontrei correspondência para \"%s\". Use Buscar imagens.", "No match found for \"%s\". Use Search images."},
	"search":             {"Buscar", "Search"},
	"close":              {"Fechar", "Close"},
	"nothing_found":      {"nada encontrado", "nothing found"},
	"sgdb_key_missing":   {"Configure a SteamGridDB API key no topo primeiro.", "Configure the SteamGridDB API key at the top first."},
	"sgdb_search_ph":     {"buscar jogo na SteamGridDB…", "search game on SteamGridDB…"},
	"sgdb_for":           {"Imagens para: %s", "Images for: %s"},
	"sgdb_showing":       {"Jogo: %s", "Game: %s"},
	"src_sgdb":           {"SteamGridDB", "SteamGridDB"},
	"src_official":       {"Steam Oficial", "Official Steam"},
	"need_match":         {"É preciso um jogo da Steam casado para ver a arte oficial. Use a busca para casar primeiro.", "A matched Steam game is required for official art. Search to match it first."},
	"official_unavail":   {"Indisponível na Steam oficial (ou ícone, que não existe na loja)", "Not available on official Steam (or icon, which doesn't exist in store)"},
	"official_log":       {"Steam oficial: arte %s aplicada ao atalho %d", "Official Steam: art %s applied to shortcut %d"},
	"search_hint":        {"Não é este jogo? Busque outro na SteamGridDB.", "Wrong game? Search another on SteamGridDB."},
	"official_for":       {"Arte oficial da Steam (app %d)", "Official Steam art (app %d)"},
	"animated_only":      {"só animadas (webp)", "animated only (webp)"},
	"no_images":          {"sem imagens", "no images"},
	"apply":              {"Aplicar", "Apply"},
	"applied":            {"Aplicado ✓", "Applied ✓"},
	"sgdb_log":           {"SGDB: arte %s aplicada ao atalho %d", "SGDB: art %s applied to shortcut %d"},
	"remove_title":       {"Limpar imagens", "Clear images"},
	"remove_confirm":     {"Remover toda a arte (capa/hero/logo/ícone) deste atalho? A arte atual será movida para grid/backup/.", "Remove all art (grid/hero/logo/icon) from this shortcut? Current art will be moved to grid/backup/."},
	"removed":            {"%d arquivo(s) movido(s) para backup.", "%d file(s) moved to backup."},
	"language":           {"Idioma:", "Language:"},
	"about":              {"Sobre", "About"},
	"about_title":        {"Sobre o SteamArt", "About SteamArt"},
	"app_name":           {"SteamArt", "SteamArt"},
	"about_desc":         {"Aplica metadados e arte (capa/hero/logo/ícone) em atalhos não-Steam da Steam, usando o catálogo oficial (CDN) ou a comunidade SteamGridDB.", "Applies metadata and art (grid/hero/logo/icon) to non-Steam Steam shortcuts, using the official Steam catalog (CDN) or the SteamGridDB community."},
	"about_links":        {"GitHub: https://github.com/SEU_USUARIO/steamart", "GitHub: https://github.com/SEU_USUARIO/steamart"},
	"tab_grid":           {"grid", "grid"},
	"tab_hero":           {"hero", "hero"},
	"tab_logo":           {"logo", "logo"},
	"tab_icon":           {"icon", "icon"},
	"tab_capsule":        {"capa horiz.", "capsule"},
}

// T traduz a chave para o idioma ativo. Argumentos opcionais são passados a
// fmt.Sprintf. Em caso de chave ausente, retorna a própria chave.
func T(key string, args ...any) string {
	e, ok := dict[key]
	if !ok {
		return key
	}
	s := e.pt
	if current == EN {
		s = e.en
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

// Languages retorna os idiomas suportados.
func Languages() []Lang { return []Lang{PTBR, EN} }

// LangName retorna o nome legível do idioma.
func LangName(l Lang) string {
	switch l {
	case PTBR:
		return "Português (BR)"
	case EN:
		return "English"
	}
	return string(l)
}

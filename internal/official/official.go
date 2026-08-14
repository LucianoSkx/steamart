// Pacote official fornece acesso às artes oficiais da loja Steam (CDN),
// equivalentes à flag --official do steamer. Ícone não existe nessa CDN,
// então não é retornado.
package official

import "fmt"

const base = "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/%d/%s"

// Asset é uma arte oficial da loja Steam para um appid.
type Asset struct {
	Kind string // grid, hero, logo, capsule
	URL  string
	Name string
}

// List retorna as artes oficiais disponíveis para um Steam appid.
func List(steamAppID int) []Asset {
	return []Asset{
		{"grid", fmt.Sprintf(base, steamAppID, "library_600x900.jpg"), "Steam · library"},
		{"hero", fmt.Sprintf(base, steamAppID, "library_hero.jpg"), "Steam · hero"},
		{"logo", fmt.Sprintf(base, steamAppID, "logo.png"), "Steam · logo"},
		{"capsule", fmt.Sprintf(base, steamAppID, "header.jpg"), "Steam · capsule"},
	}
}

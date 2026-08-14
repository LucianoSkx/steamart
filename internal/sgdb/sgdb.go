package sgdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const base = "https://www.steamgriddb.com/api/v2"

var client = &http.Client{Timeout: 30 * time.Second}

// Client é um cliente da API SteamGridDB.
type Client struct {
	APIKey string
}

func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// Game é um jogo retornado pela busca.
type Game struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Image é uma imagem da comunidade.
type Image struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Thumb  string `json:"thumb"`
	Author struct {
		Name string `json:"name"`
	} `json:"author"`
	Style      string `json:"style"`
	Dimensions string `json:"dimensions"`
	NSFW       bool   `json:"nsfw"`
}

// DownloadURL retorna a URL a baixar. Para ícones, prefere o thumb (PNG),
// já que muitos ícones da SGDB vêm em .ico e o thumb é PNG quadrado.
func (i Image) DownloadURL(asset string) string {
	if asset == "icon" && i.Thumb != "" {
		return i.Thumb
	}
	return i.URL
}

func (c *Client) do(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", "steamart/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SteamGridDB status %d: %s", resp.StatusCode, string(body))
	}
	var wrapper struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return err
	}
	if !wrapper.Success {
		return fmt.Errorf("SteamGridDB retornou sucesso=false")
	}
	return json.Unmarshal(wrapper.Data, out)
}

// Search busca jogos pelo termo.
func (c *Client) Search(term string) ([]Game, error) {
	term = url.PathEscape(term)
	var games []Game
	if err := c.do("/search/autocomplete/"+term, &games); err != nil {
		return nil, err
	}
	return games, nil
}

// GameBySteamAppID encontra o jogo do SGDB correspondente a um Steam appid.
func (c *Client) GameBySteamAppID(steamAppID int) (*Game, error) {
	var g Game
	if err := c.do(fmt.Sprintf("/games/steam/%d", steamAppID), &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// Images lista imagens de um tipo (grid, hero, logo, icon, capsule) para um
// jogo. dims filtra por dimensão (ex.: "920x430" para capa horizontal); quando
// animated=true, filtra por formatos animados (webp/gif).
// A API da SteamGridDB usa nomes no plural (grids/heroes/logos/icons).
func (c *Client) Images(gameID int, asset, dims string, animated bool) ([]Image, error) {
	ep := asset + "s"
	switch asset {
	case "hero":
		ep = "heroes"
	case "capsule":
		ep = "grids"
	}
	path := fmt.Sprintf("/%s/game/%d", ep, gameID)
	q := url.Values{}
	if dims != "" {
		q.Set("dimensions", dims)
	}
	if animated {
		q.Set("mimes", "image/webp,image/gif")
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var images []Image
	if err := c.do(path, &images); err != nil {
		return nil, err
	}
	return images, nil
}

// AssetTypes são os tipos de arte suportados.
var AssetTypes = []string{"grid", "hero", "logo", "icon"}

# Contribuindo

Obrigado por querer contribuir com o SteamArt! Esta guia explica como
configurar o ambiente e enviar suas mudanças.

## Pré-requisitos

- Go 1.26+ (ou `go 1.26` como no `go.mod`)
- gcc + dev libs do Fyne (GTK3, GL, X11) — veja a seção "Pré-requisitos" no README

## Setup rápido

```bash
git clone https://github.com/LucianoSkx/steamart.git
cd steamart
make test     # roda todos os testes
make          # compila ./steamart
```

## Fluxo de contribuição

1. **Fork** o repositório
2. **Branch** com uma feature clara:
   ```bash
   git checkout -b feature/minha-feature
   ```
3. **Codifique** seguindo as regras:
   - `gofmt -w .` antes de commitar (formatação obrigatória) — use `make fmt`
   - `go vet ./...` sem warnings — use `make vet`
   - `go test ./...` verde — use `make test`
   - Comentários em **português** (pt-BR)
4. **Commit** com mensagem clara:
   ```bash
   git commit -m "feat: aplicar arte X no card"
   ```
5. **Push** & **Pull Request**:
   ```bash
   git push origin feature/minha-feature
   ```
   Abra um PR descrevendo a motivação e os testes feitos.

## Diretrizes de código

- Siga a estrutura existente: `internal/` para lógica, `cmd/` para entrypoints
- Funções pequenas, nomes claros, comentários de pacote em pt-BR
- Evite dependências externas não listadas no `go.mod`
- Testes unitários para novas funcionalidades (especialmente em `internal/match`,
  `internal/title`, `internal/artwork`)

## Build de AppImage (local)

```bash
bash build-appimage.sh      # gera SteamArt-<versao>-x86_64.AppImage
```

Precisa do `linuxdeploy` + `wget`. O script baixa automaticamente se ausente.

## Releases

Releases são publicados **automaticamente** pelo GitHub Actions a cada tag `v*`
(empacotamento `git tag v1.2.3 && git push --tags`). Não é preciso abrir PR para
release — apenas faça a tag após mergear para a `main`.

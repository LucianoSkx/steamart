package main

import (
	"os"

	"steamart/internal/icon"
)

func main() {
	b, err := icon.PNG(512)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("assets/icon.png", b, 0o644); err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.ExpandEnv("$HOME/.local/share/icons/steamart.png"), b, 0o644); err != nil {
		panic(err)
	}
	println("ícone gerado: assets/icon.png e ~/.local/share/icons/steamart.png")
}

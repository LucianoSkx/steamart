package main

import (
	"fmt"
	"os"

	"steamart/internal/vdf"
)

func typ(v any) string {
	switch v.(type) {
	case *vdf.Node:
		return "dict"
	case string:
		return "str"
	case int32:
		return "int"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	root, err := vdf.Parse(data)
	if err != nil {
		panic(err)
	}
	sc := root.Dict("shortcuts")
	if sc == nil {
		fmt.Println("no shortcuts")
		return
	}
	fmt.Println("shortcuts children:", len(sc.Children))
	for _, c := range sc.Children {
		fmt.Printf("  key=%-4s type=%s\n", c.Key, typ(c.Value))
		if d, ok := c.Value.(*vdf.Node); ok {
			appid := uint32(d.Int("appid"))
			fmt.Printf("    appid=%d (0x%X) name=%q\n", appid, appid, d.Str("AppName"))
		}
	}
}

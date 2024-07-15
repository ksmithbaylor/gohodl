package main

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/config"
)

func main() {
	fmt.Printf("Config: %#+v\n", config.Config)
}

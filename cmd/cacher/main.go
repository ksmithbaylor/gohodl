package main

import (
	"fmt"
	"log"

	"github.com/ksmithbaylor/gohodl/internal/util"
)

type thingy struct {
	A int    `json:"a"`
	B string `json:"b"`
	C bool   `json:"c"`
}

func main() {
	cache, err := util.NewFileCache("stuff")
	if err != nil {
		log.Fatal(err)
	}

	var foo []string
	fooFound, err := cache.Read("foo", &foo)
	if err != nil {
		log.Fatal(err)
	}
	if fooFound {
		fmt.Println("Found foo!")
		fmt.Println(foo)
	} else {
		fmt.Println("Writing foo as a list of strings")
		err := cache.Write("foo", []string{"one", "two", "three"})
		if err != nil {
			log.Fatal(err)
		}
	}

	var bar thingy
	barFound, err := cache.Read("bar", &bar)
	if err != nil {
		log.Fatal(err)
	}
	if barFound {
		fmt.Println("Found bar!")
		fmt.Println(bar)
	} else {
		fmt.Println("Writing bar as a struct with fields")
		err := cache.Write("bar", thingy{42, "asdf", true})
		if err != nil {
			log.Fatal(err)
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/slimcdk/go-police-news"
)

func prettyPrint(emp ...interface{}) {
	empJSON, err := json.MarshalIndent(emp, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println(string(empJSON))
}

func main() {

	p := police.New()

}

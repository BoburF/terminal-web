package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/BoburF/terminal-web/engine/ml"
)

func main() {
	fmt.Println("Welcome to TWeb")

	const code = `                          basbdasd
		asbasbdaksdb /block/ bad boy /end/
		// / // / /    /`

	reader := strings.NewReader(code)

	tokenizer := ml.NewTokenizer(reader)

	for {
		token, err := tokenizer.Next()
		if err != nil {
			if err == io.EOF {
				fmt.Println(token, "end")
				break
			}
			fmt.Println("err:", token, err)
		}

		fmt.Println(token)
	}
}

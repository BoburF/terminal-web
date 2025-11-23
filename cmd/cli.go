package main

import (
	"fmt"
	"strings"

	"github.com/BoburF/terminal-web/engine/ml"
)

func main() {
	fmt.Println("Welcome to TWeb")

	const code = `                          basbdasd
		asbasbdaksdb /block/ what up bad boy /end/
		// / // / /    //block/ hey what up I said!/end/`

	reader := strings.NewReader(code)

	tokenizer := ml.NewTokenizer(reader)

	parser := ml.NewParser()

	ast, err := parser.Analyze(tokenizer)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	for _, node := range ast.Children {
		fmt.Println(node.Type, node.Value)
	}
}

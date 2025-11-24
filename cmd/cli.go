package main

import (
	"fmt"
	"strings"

	"github.com/BoburF/terminal-web/engine/ml"
)

func main() {
	fmt.Println("Welcome to TWeb")

	const code = `                          basbdasd
		asbasbdaksdb /block/ what up bad boy /block/ Bobur /end/ /end/
		// / // / /    //block/ hey what up I said!/end/`

	reader := strings.NewReader(code)

	tokenizer := ml.NewTokenizer(reader)

	parser := ml.NewParser()

	ast, err := parser.Analyze(tokenizer)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	GoThrough(ast, 0)
}

func GoThrough(ast ml.Node, tabs int) {
	for _, node := range ast.Children {
		fmt.Println(strings.Repeat(" ", tabs), node.Type, node.Value)
		if len(node.Children) > 0 {
			newTabs := tabs + 1
			GoThrough(node, newTabs)
		}
	}
}

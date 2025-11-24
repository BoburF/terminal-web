package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BoburF/terminal-web/engine/ml"
)

func main() {
	pathFile := flag.String("path", "/path/file.ml", "path to read the file")

	flag.Parse()

	file, err := os.Open(*pathFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tokenizer := ml.NewTokenizer(file)

	parser := ml.NewParser()

	ast, err := parser.Analyze(tokenizer)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	dir := filepath.Dir(*pathFile)
	base := filepath.Base(*pathFile)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	newPath := filepath.Join(dir, nameWithoutExt+".json")

	astJSON, err := json.Marshal(ast)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(newPath, astJSON, 0o644)
	if err != nil {
		log.Fatal(err)
	}
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

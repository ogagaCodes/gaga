package compiler

import (
	"os"
	"strings"
)

type Compiler struct {
	SourceFile string
	OutputFile string
}

func NewCompiler(source, output string) *Compiler {
	return &Compiler{
		SourceFile: source,
		OutputFile: output,
	}
}

func (c *Compiler) Compile() error {
	content, err := os.ReadFile(c.SourceFile)
	if err != nil {
		return err
	}

	lexer := NewLexer(strings.NewReader(string(content)))
	parser := NewParser(lexer)

	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	return CompileToGo(ast, c.OutputFile)
}

package compiler

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CodeGenerator struct {
	ast *AST
}

func (c *CodeGenerator) Generate() (string, error) {
	code := "package main\n\nimport (\n\t\"fmt\"\n"

	// Add imports
	for _, imp := range c.ast.Imports {
		code += fmt.Sprintf("\t. \"%s\"\n", imp)
	}
	code += ")\n\n"

	// Generate functions
	for _, stmt := range c.ast.Statements {
		switch s := stmt.(type) {
		case *ArrayDeclaration:
			code += fmt.Sprintf("var %s = []interface{}{", s.Name)
			for i, elem := range s.Elements {
				if i > 0 {
					code += ", "
				}
				switch v := elem.(type) {
				case string:
					if _, err := strconv.Atoi(v); err == nil {
						code += v // Unquoted number string
					} else {
						code += fmt.Sprintf("%q", v)
					}
				case int:
					code += fmt.Sprintf("%d", v)
				case float64:
					code += fmt.Sprintf("%f", v)
				}
			}
			code += "}\n"
		case *FunctionDecl:
			code += fmt.Sprintf("func %s(%s) interface{} {\n",
				s.Name, strings.Join(s.Params, ", "))
			code += "\t// Auto-generated function\n"
			code += "}\n\n"
		case *VarStatement:
			code += fmt.Sprintf("var %s = %s\n", s.Name, s.Value)
		case *DoStatement:
			code += "func main() {\n"
			code += fmt.Sprintf("\tfmt.Println(%s)\n", strings.Join(toStrings(s.Args), " "))
			code += "}\n"
		}
	}

	return code, nil
}

func toStrings(args []interface{}) []string {
	var res []string
	for _, a := range args {
		res = append(res, fmt.Sprint(a))
	}
	return res
}

func CompileToGo(ast *AST, outputPath string) error {
	gen := &CodeGenerator{ast: ast}
	code, err := gen.Generate()
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(code), 0644)
}

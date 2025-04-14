package compiler

type Parser struct {
	lexer   *Lexer
	current Token
}

type AST struct {
	Imports    []string
	Statements []Statement
}

type Statement interface{}

type DoStatement struct {
	Operation string
	Args      []interface{}
}

type FunctionDecl struct {
	Name   string
	Params []string
	Body   []Statement
}

type VarStatement struct {
	Name  string
	Value interface{}
}

type ImportStatement struct {
	Path string
}

func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer}
}

func (p *Parser) Parse() (*AST, error) {
	ast := &AST{}

	for {
		tok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}
		if tok.Type == TOKEN_EOF {
			break
		}

		switch tok.Type {
		case TOKEN_DO:
			stmt, err := p.parseDo()
			if err != nil {
				return nil, err
			}
			ast.Statements = append(ast.Statements, stmt)
		case TOKEN_MAKE:
			stmt, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			ast.Statements = append(ast.Statements, stmt)
		case TOKEN_GET:
			pathTok, err := p.lexer.Lex()
			if err != nil {
				return nil, err
			}
			ast.Imports = append(ast.Imports, pathTok.Value)
		case TOKEN_IDENT:
			if tok.Value == "variable" {
				stmt, err := p.parseVar()
				if err != nil {
					return nil, err
				}
				ast.Statements = append(ast.Statements, stmt)
			}
		}
	}

	return ast, nil
}

func (p *Parser) parseDo() (*DoStatement, error) {
	opTok, err := p.lexer.Lex()
	if err != nil {
		return nil, err
	}

	stmt := &DoStatement{Operation: opTok.Value}

	for {
		tok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}
		if tok.Type == TOKEN_EOF || tok.Type == TOKEN_DO {
			p.unread(tok)
			break
		}
		stmt.Args = append(stmt.Args, tok.Value)
	}

	return stmt, nil
}

func (p *Parser) parseFunction() (*FunctionDecl, error) {
	fn := &FunctionDecl{}

	// Parse function name
	nameTok, err := p.lexer.Lex()
	if err != nil {
		return nil, err
	}
	fn.Name = nameTok.Value

	// Parse parameters
	for {
		tok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}
		if tok.Type == TOKEN_END {
			break
		}
		if tok.Type == TOKEN_IDENT {
			fn.Params = append(fn.Params, tok.Value)
		}
	}

	return fn, nil
}

func (p *Parser) parseVar() (*VarStatement, error) {
	nameTok, err := p.lexer.Lex()
	if err != nil {
		return nil, err
	}

	_, err = p.lexer.Lex() // consume 'is'
	if err != nil {
		return nil, err
	}

	valTok, err := p.lexer.Lex()
	if err != nil {
		return nil, err
	}

	return &VarStatement{
		Name:  nameTok.Value,
		Value: valTok.Value,
	}, nil
}

func (p *Parser) unread(tok Token) {
	// Simplified unread for this implementation
}

package compiler

import (
	"fmt"
	"strconv"
)

type Parser struct {
	lexer   *Lexer
	current Token
}

type ArrayDeclaration struct {
	Name     string
	Elements []interface{}
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

type NetworkCall struct {
	Method   string
	URL      string
	Payload  string
	Parallel bool
}

type ParallelBlock struct {
	Calls []*NetworkCall
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
		case TOKEN_PARALLEL:
			parallelBlock, err := p.parseParallel()
			if err != nil {
				return nil, err
			}
			ast.Statements = append(ast.Statements, parallelBlock)
		case TOKEN_POST, TOKEN_GETCONNECTION, TOKEN_PUT, TOKEN_DELETE:
			call, err := p.parseNetworkCall(tok.Type)
			if err != nil {
				return nil, err
			}
			ast.Statements = append(ast.Statements, call)
		case TOKEN_DO:
			stmt, err := p.parseDo()
			if err != nil {
				return nil, err
			}
			ast.Statements = append(ast.Statements, stmt)
		case TOKEN_MAKE:
			// Peek next token to determine make type
			nextTok, err := p.lexer.Lex()
			if err != nil {
				return nil, err
			}
			p.unread(nextTok) // Put back the token we peeked

			switch nextTok.Type {
			case TOKEN_ARRAY:
				arr, err := p.parseArray()
				if err != nil {
					return nil, err
				}
				ast.Statements = append(ast.Statements, arr)
			case TOKEN_FUNCTION:
				stmt, err := p.parseFunction()
				if err != nil {
					return nil, err
				}
				ast.Statements = append(ast.Statements, stmt)
			default:
				return nil, fmt.Errorf("unexpected token after MAKE: %v", nextTok)
			}
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

func (p *Parser) parseArray() (*ArrayDeclaration, error) {
	// Get array name
	nameTok, err := p.lexer.Lex()
	if err != nil || nameTok.Type != TOKEN_IDENT {
		return nil, fmt.Errorf("expected array name")
	}

	arr := &ArrayDeclaration{Name: nameTok.Value}

	// Parse elements
	for {
		tok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}
		if tok.Type == TOKEN_NEWLINE || tok.Type == TOKEN_EOF {
			break
		}

		switch tok.Type {
		case TOKEN_IDENT:
			arr.Elements = append(arr.Elements, tok.Value)
		case TOKEN_NUMBER:
			num, _ := strconv.Atoi(tok.Value)
			arr.Elements = append(arr.Elements, num)
		case TOKEN_FLOAT:
			num, _ := strconv.ParseFloat(tok.Value, 64)
			arr.Elements = append(arr.Elements, num)
		case TOKEN_STRING:
			arr.Elements = append(arr.Elements, tok.Value)
		default:
			return nil, fmt.Errorf("invalid array element: %v", tok)
		}
	}

	return arr, nil
}

func (p *Parser) unread(tok Token) {
	// Simplified unread for this implementation
}

func (p *Parser) parseNetworkCall(method TokenType) (*NetworkCall, error) {
	call := &NetworkCall{}

	// Get method
	switch method {
	case TOKEN_POST:
		call.Method = "POST"
	case TOKEN_GET:
		call.Method = "GET"
	case TOKEN_PUT:
		call.Method = "PUT"
	case TOKEN_DELETE:
		call.Method = "DELETE"
	}

	// Parse URL
	_, err := p.lexer.Lex() // consume 'to', 'from', or 'in'
	urlTok, err := p.lexer.Lex()
	if err != nil {
		return nil, err
	}
	call.URL = urlTok.Value

	// Parse payload if exists
	payloadTok, err := p.lexer.Lex()
	if err == nil && payloadTok.Type == TOKEN_PAYLOAD {
		dataTok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}
		call.Payload = dataTok.Value
	} else {
		p.unread(payloadTok)
	}

	return call, nil
}

func (p *Parser) parseParallel() (*ParallelBlock, error) {
	block := &ParallelBlock{}

	for {
		tok, err := p.lexer.Lex()
		if err != nil {
			return nil, err
		}

		if tok.Type == TOKEN_END {
			break
		}

		if tok.Type == TOKEN_POST || tok.Type == TOKEN_GET ||
			tok.Type == TOKEN_PUT || tok.Type == TOKEN_DELETE {
			call, err := p.parseNetworkCall(tok.Type)
			if err != nil {
				return nil, err
			}
			call.Parallel = true
			block.Calls = append(block.Calls, call)
		}
	}

	return block, nil
}

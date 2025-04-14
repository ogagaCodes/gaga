package compiler

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type TokenType int

const (
	TOKEN_EOF TokenType = iota
	TOKEN_DO
	TOKEN_MAKE
	TOKEN_FUNCTION
	TOKEN_SUM
	TOKEN_SUBTRACT
	TOKEN_IS
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_IDENT
	TOKEN_GET
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_ARRAY
	TOKEN_FLOAT
	TOKEN_NEWLINE
	TOKEN_END
	TOKEN_POST
	TOKEN_GETCONNECTION
	TOKEN_PUT
	TOKEN_DELETE
	TOKEN_TO
	TOKEN_FROM
	TOKEN_IN
	TOKEN_PAYLOAD
	TOKEN_PARALLEL
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	r *bufio.Reader
}

const eof = rune(0)

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{r: bufio.NewReader(r)}
}

func (l *Lexer) read() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (l *Lexer) unread() { _ = l.r.UnreadRune() }

func (l *Lexer) Lex() (Token, error) {
	for {
		ch := l.read()

		if ch == eof {
			return Token{TOKEN_EOF, ""}, nil
		}

		if unicode.IsSpace(ch) {
			continue
		}

		if unicode.IsDigit(ch) {
			l.unread()
			return l.lexNumber()
		}

		if ch == '"' {
			l.unread()
			return l.lexString()
		}

		if unicode.IsLetter(ch) {
			l.unread()
			return l.lexIdent()
		}

		if ch == '(' {
			return Token{TOKEN_LPAREN, "("}, nil
		}

		if ch == ')' {
			return Token{TOKEN_RPAREN, ")"}, nil
		}
	}
}

func (l *Lexer) lexNumber() (Token, error) {
	var buf strings.Builder
	isFloat := false

	for {
		ch := l.read()
		if ch == eof {
			break
		}

		if ch == '.' && !isFloat {
			isFloat = true
			buf.WriteRune(ch)
		} else if !unicode.IsDigit(ch) {
			l.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	if isFloat {
		return Token{TOKEN_FLOAT, buf.String()}, nil
	}
	return Token{TOKEN_NUMBER, buf.String()}, nil
}

func (l *Lexer) lexString() (Token, error) {
	l.read() // consume opening "
	var buf strings.Builder

	for {
		ch := l.read()
		if ch == '"' || ch == eof {
			break
		}
		buf.WriteRune(ch)
	}

	return Token{TOKEN_STRING, buf.String()}, nil
}

func (l *Lexer) lexIdent() (Token, error) {
	var buf strings.Builder
	for {
		ch := l.read()
		if ch == eof {
			break
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			l.unread()
			break
		}
		buf.WriteRune(ch)
	}

	str := strings.ToLower(buf.String())
	switch str {
	case "do":
		return Token{TOKEN_DO, str}, nil
	case "make":
		return Token{TOKEN_MAKE, str}, nil
	case "function":
		return Token{TOKEN_FUNCTION, str}, nil
	case "sum":
		return Token{TOKEN_SUM, str}, nil
	case "subtract":
		return Token{TOKEN_SUBTRACT, str}, nil
	case "is":
		return Token{TOKEN_IS, str}, nil
	case "get":
		return Token{TOKEN_GET, str}, nil
	case "end":
		return Token{TOKEN_END, str}, nil
	case "array":
		return Token{TOKEN_ARRAY, str}, nil

	case "post":
		return Token{TOKEN_POST, str}, nil
	case "getConnection":
		return Token{TOKEN_GETCONNECTION, str}, nil
	case "put":
		return Token{TOKEN_PUT, str}, nil
	case "delete":
		return Token{TOKEN_DELETE, str}, nil
	case "to":
		return Token{TOKEN_TO, str}, nil
	case "from":
		return Token{TOKEN_FROM, str}, nil
	case "in":
		return Token{TOKEN_IN, str}, nil
	case "payload":
		return Token{TOKEN_PAYLOAD, str}, nil
	case "parallel":
		return Token{TOKEN_PARALLEL, str}, nil
	default:
		return Token{TOKEN_IDENT, buf.String()}, nil
	}
}

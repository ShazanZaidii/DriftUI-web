package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	STRING = "STRING"
	INT    = "INT"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
	ASSIGN   = "="
	AT       = "@"
	COMMA    = ","
	DOT      = "."
	COLON    = ":"

	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="
	OR       = "||"
	AND      = "&&"
	LTE      = "<="
	GTE      = ">="

	FUN   = "FUN"
	APP   = "APP"
	SET   = "SET"
	STATE = "State"
	VAR   = "VAR"
	VAL   = "VAL"
	TRUE  = "TRUE"
	FALSE = "FALSE"
	IF    = "IF"
	ELSE  = "ELSE"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"fun":   FUN,
	"App":   APP,
	"set":   SET,
	"State": STATE,
	"var":   VAR,
	"val":   VAL,
	"true":  TRUE,
	"false": FALSE,
	"if":    IF,
	"else":  ELSE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
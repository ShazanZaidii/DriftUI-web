package parser

import (
	"drift/ast"
	"drift/lexer"
	"drift/token"
	"fmt"
)

type ParseError struct {
	Message string
	Line    int
	Column  int
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	Errors    []ParseError
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, Errors: []ParseError{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) addError(msg string, t token.Token) {
	p.Errors = append(p.Errors, ParseError{Message: msg, Line: t.Line, Column: t.Column})
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("Expected next token to be '%s', got '%s'", t, p.peekToken.Type), p.peekToken)
	return false
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseFile() *ast.File {
	file := &ast.File{Components: []*ast.Component{}}
	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.FUN {
			if comp := p.parseComponent(); comp != nil { file.Components = append(file.Components, comp) } else { p.nextToken() }
		} else if p.curToken.Type == token.APP {
			if app := p.parseAppRoot(); app != nil { file.AppRoot = app } else { p.nextToken() }
		} else {
			p.nextToken()
		}
	}
	return file
}

func (p *Parser) parseAppRoot() *ast.AppRoot {
	if !p.expectPeek(token.LBRACE) { return nil }
	p.nextToken()
	entryName := p.curToken.Literal
	p.nextToken()
	if p.curToken.Type == token.LPAREN {
		p.nextToken()
		p.nextToken()
	}
	p.nextToken()
	return &ast.AppRoot{EntryComponent: entryName}
}

func (p *Parser) parseComponent() *ast.Component {
	comp := &ast.Component{Metadata: make(map[string]string), Declarations: []ast.Node{}}
	if !p.expectPeek(token.IDENT) { return nil }
	comp.Name = p.curToken.Literal

	if comp.Name[0] >= 'a' && comp.Name[0] <= 'z' {
		p.addError(fmt.Sprintf("Component declarations must start with an uppercase letter: '%s'", comp.Name), p.curToken)
	}

	if !p.expectPeek(token.LPAREN) { return nil }
	if !p.expectPeek(token.RPAREN) { return nil }
	if !p.expectPeek(token.LBRACE) { return nil }
	p.nextToken()

	for p.curToken.Literal == "set" {
		p.nextToken()
		key := p.curToken.Literal
		p.nextToken()
		p.nextToken()
		comp.Metadata[key] = p.curToken.Literal
		p.nextToken()
	}

	for p.curToken.Type == token.AT || p.curToken.Type == token.VAL || p.curToken.Type == token.VAR {
		var decl ast.Node
		if p.curToken.Type == token.AT {
			decl = p.parseStateDeclaration()
		} else {
			decl = p.parseVariableDeclaration()
		}
		if decl != nil { comp.Declarations = append(comp.Declarations, decl) } else { p.nextToken() }
	}
	comp.UI = p.parseNode()
	p.nextToken()
	return comp
}

func (p *Parser) parseStateDeclaration() *ast.StateDeclaration {
	if !p.expectPeek(token.STATE) { return nil }
	if !p.expectPeek(token.VAR) { return nil }
	if !p.expectPeek(token.IDENT) { return nil }
	name := p.curToken.Literal
	if !p.expectPeek(token.ASSIGN) { return nil }
	p.nextToken()
	val := p.parseExpression()
	return &ast.StateDeclaration{Name: name, Value: val}
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	isVal := p.curToken.Type == token.VAL
	if !p.expectPeek(token.IDENT) { return nil }
	name := p.curToken.Literal
	if !p.expectPeek(token.ASSIGN) { return nil }
	p.nextToken()
	val := p.parseExpression()
	return &ast.VariableDeclaration{IsVal: isVal, Name: name, Value: val}
}

func isOperator(t token.TokenType) bool {
	return t == token.PLUS || t == token.MINUS || t == token.ASTERISK || t == token.SLASH ||
		t == token.LT || t == token.GT || t == token.EQ || t == token.NOT_EQ ||
		t == token.OR || t == token.AND || t == token.LTE || t == token.GTE
}

func (p *Parser) parseExpression() ast.Node {
	left := p.parseSingleExpression()
	for isOperator(p.curToken.Type) {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseSingleExpression()
		left = &ast.InfixExpression{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseSingleExpression() ast.Node {
	var node ast.Node
	switch p.curToken.Type {
	case token.STRING:
		node = &ast.StringLiteral{Value: p.curToken.Literal}
		p.nextToken()
	case token.INT:
		node = &ast.IntLiteral{Value: p.curToken.Literal}
		p.nextToken()
		if p.curToken.Type == token.DOT && (p.peekToken.Literal == "dp" || p.peekToken.Literal == "sp" || p.peekToken.Literal == "f") {
			p.nextToken()
			p.nextToken()
		}
	case token.TRUE:
		node = &ast.BooleanLiteral{Value: true}
		p.nextToken()
	case token.FALSE:
		node = &ast.BooleanLiteral{Value: false}
		p.nextToken()
	case token.LBRACKET:
		node = p.parseArray()
	case token.LBRACE:
		node = p.parseDictionary()
	case token.IF:
		node = p.parseIfExpression()
	case token.IDENT:
		node = &ast.Identifier{Value: p.curToken.Literal}
		p.nextToken()
	}

	for p.curToken.Type == token.DOT || p.curToken.Type == token.LPAREN {
		if p.curToken.Type == token.DOT {
			p.nextToken()
			propName := p.curToken.Literal
			p.nextToken()

			if p.curToken.Type == token.LPAREN {
				args := p.parseArguments()
				var block ast.Node
				if p.curToken.Type == token.LBRACE { block = p.parseBlockLiteral() }
				node = &ast.MethodCall{CallerNode: node, Method: propName, Arguments: args, Block: block}
			} else if p.curToken.Type == token.LBRACE {
				block := p.parseBlockLiteral()
				node = &ast.MethodCall{CallerNode: node, Method: propName, Arguments: []ast.Argument{}, Block: block}
			} else {
				node = &ast.PropertyAccess{ObjectNode: node, Property: propName}
			}
		} else if p.curToken.Type == token.LPAREN {
			args := p.parseArguments()
			var block ast.Node
			if p.curToken.Type == token.LBRACE { block = p.parseBlockLiteral() }
			node = &ast.FunctionCall{CallerNode: node, Arguments: args, Block: block}
		}
	}
	return node
}

func (p *Parser) parseArguments() []ast.Argument {
	args := []ast.Argument{}
	p.nextToken()
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		var name string
		isNamed := false
		line := p.curToken.Line
		col := p.curToken.Column

		if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN {
			name = p.curToken.Literal
			isNamed = true
			p.nextToken()
			p.nextToken()
			line = p.curToken.Line
			col = p.curToken.Column
		}
		var val ast.Node
		if p.curToken.Type == token.LBRACE {
			val = p.parseBlockLiteral()
		} else {
			val = p.parseExpression()
		}
		if val != nil {
			if name == "" { name = fmt.Sprintf("arg%d", len(args)) }
			args = append(args, ast.Argument{Name: name, Value: val, IsNamed: isNamed, Line: line, Column: col})
		} else {
			p.nextToken()
		}
		if p.curToken.Type == token.COMMA { p.nextToken() }
	}
	if p.curToken.Type == token.RPAREN { p.nextToken() }
	return args
}

func (p *Parser) parseModifierChain() *ast.ModifierChain {
	chain := &ast.ModifierChain{Calls: []ast.ModifierCall{}}
	p.nextToken() // Escapes "Modifier" identifier

	for p.curToken.Type == token.DOT {
		p.nextToken() // Escapes DOT
		if p.curToken.Type != token.IDENT { break }
		call := ast.ModifierCall{Name: p.curToken.Literal, Arguments: []ast.Node{}}
		p.nextToken() // Escapes Modifier Name

		if p.curToken.Type == token.LPAREN {
			p.nextToken()
			for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
				if p.curToken.Type != token.COMMA {
					if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN {
						p.nextToken()
						p.nextToken()
					}
					if val := p.parseExpression(); val != nil { call.Arguments = append(call.Arguments, val) } else { p.nextToken() }
				}
				if p.curToken.Type == token.COMMA { p.nextToken() }
			}
			if p.curToken.Type == token.RPAREN { p.nextToken() }
		}

		if p.curToken.Type == token.LBRACE {
			block := p.parseBlockLiteral()
			call.Arguments = append(call.Arguments, block)
		}
		chain.Calls = append(chain.Calls, call)
	}
	return chain
}

func (p *Parser) parseBlockLiteral() *ast.BlockLiteral {
	block := &ast.BlockLiteral{Statements: []ast.Node{}}
	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN {
			name := p.curToken.Literal
			p.nextToken()
			p.nextToken()
			val := p.parseExpression()
			block.Statements = append(block.Statements, &ast.Assignment{Name: name, Value: val})
		} else {
			expr := p.parseExpression()
			if expr != nil { block.Statements = append(block.Statements, expr) } else { p.nextToken() }
		}
	}
	if p.curToken.Type == token.RBRACE { p.nextToken() }
	return block
}

func (p *Parser) parseIfExpression() *ast.IfExpression {
	p.nextToken()
	p.nextToken()
	condition := p.parseExpression()
	if p.curToken.Type == token.RPAREN { p.nextToken() } else { p.nextToken() }

	var block ast.Node
	if p.curToken.Type == token.LBRACE {
		block = p.parseBlockLiteral()
	} else {
		block = p.parseExpression()
	}

	var alternative ast.Node
	if p.curToken.Type == token.ELSE {
		p.nextToken()
		if p.curToken.Type == token.IF {
			alternative = p.parseIfExpression()
		} else if p.curToken.Type == token.LBRACE {
			alternative = p.parseBlockLiteral()
		} else {
			alternative = p.parseExpression()
		}
	}
	return &ast.IfExpression{Condition: condition, Block: block, Alternative: alternative}
}

func (p *Parser) parseComponentCall() *ast.ComponentCall {
	comp := &ast.ComponentCall{Name: p.curToken.Literal, Arguments: []ast.Argument{}}
	if comp.Name[0] >= 'a' && comp.Name[0] <= 'z' {
		p.addError(fmt.Sprintf("Invalid UI node '%s'. Custom components must be Capitalized.", comp.Name), p.curToken)
	}
	p.nextToken()

	if p.curToken.Type == token.LPAREN {
		p.nextToken()
		for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
			var argName string
			isNamed := false
			line := p.curToken.Line
			col := p.curToken.Column

			if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN {
				argName = p.curToken.Literal
				isNamed = true
				p.nextToken()
				p.nextToken()
				line = p.curToken.Line
				col = p.curToken.Column
			}
			
			// Intelligently trap Modifier whether it's named or positional
			if argName == "modifier" || (!isNamed && p.curToken.Literal == "Modifier") {
				comp.Modifier = p.parseModifierChain()
			} else {
				var argVal ast.Node
				if p.curToken.Type == token.LBRACE {
					argVal = p.parseBlockLiteral()
				} else {
					argVal = p.parseExpression()
				}
				if argVal != nil {
					if argName == "" { argName = fmt.Sprintf("arg%d", len(comp.Arguments)) }
					comp.Arguments = append(comp.Arguments, ast.Argument{Name: argName, Value: argVal, IsNamed: isNamed, Line: line, Column: col})
				} else {
					p.nextToken()
				}
			}
			if p.curToken.Type == token.COMMA { p.nextToken() }
		}
		if p.curToken.Type == token.RPAREN { p.nextToken() }
	}

	if p.curToken.Type == token.LBRACE {
		p.nextToken()
		for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
			if node := p.parseNode(); node != nil { comp.Children = append(comp.Children, node) } else { p.nextToken() }
		}
		if p.curToken.Type == token.RBRACE { p.nextToken() }
	}
	return comp
}

func (p *Parser) parseNode() ast.Node {
	switch p.curToken.Literal {
	case "Column": return p.parseContainer(func(mod *ast.ModifierChain, children []ast.Node) ast.Node { return &ast.Column{Modifier: mod, Children: children} })
	case "Row": return p.parseContainer(func(mod *ast.ModifierChain, children []ast.Node) ast.Node { return &ast.Row{Modifier: mod, Children: children} })
	case "Box": return p.parseContainer(func(mod *ast.ModifierChain, children []ast.Node) ast.Node { return &ast.Box{Modifier: mod, Children: children} })
	case "Text": return p.parseText()
	}
	if p.curToken.Type == token.IF { return p.parseIfExpression() }
	if p.curToken.Type == token.IDENT {
		if p.curToken.Literal >= "A" && p.curToken.Literal <= "Z" && p.peekToken.Type != token.DOT {
			return p.parseComponentCall()
		}
		return p.parseExpression()
	}
	return nil
}

func (p *Parser) parseContainer(constructor func(*ast.ModifierChain, []ast.Node) ast.Node) ast.Node {
	var mod *ast.ModifierChain
	p.nextToken()
	if p.curToken.Type == token.LPAREN {
		p.nextToken()
		for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
			if p.curToken.Type == token.IDENT && p.curToken.Literal == "modifier" && p.peekToken.Type == token.ASSIGN {
				p.nextToken()
				p.nextToken()
				if p.curToken.Literal == "Modifier" { mod = p.parseModifierChain() }
			} else if p.curToken.Literal == "Modifier" {
				mod = p.parseModifierChain()
			} else if p.curToken.Type != token.COMMA {
				p.nextToken()
			}
			if p.curToken.Type == token.COMMA { p.nextToken() }
		}
		if p.curToken.Type == token.RPAREN { p.nextToken() }
	}
	
	if p.curToken.Type != token.LBRACE {
		p.addError("Expected '{' for layout container", p.curToken)
		p.nextToken()
		return nil
	}
	p.nextToken()
	
	var children []ast.Node
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if node := p.parseNode(); node != nil { children = append(children, node) } else { p.nextToken() }
	}
	if p.curToken.Type == token.RBRACE { p.nextToken() }
	return constructor(mod, children)
}

func (p *Parser) parseText() *ast.Text {
	node := &ast.Text{}
	p.nextToken()
	p.nextToken()
	node.Expression = p.parseExpression()
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		if (p.curToken.Type == token.IDENT && p.curToken.Literal == "modifier" && p.peekToken.Type == token.ASSIGN) || p.curToken.Literal == "Modifier" {
			if p.curToken.Literal == "modifier" {
				p.nextToken()
				p.nextToken()
			}
			if p.curToken.Literal == "Modifier" { node.Modifier = p.parseModifierChain() }
		} else if p.curToken.Type != token.COMMA { p.nextToken() }
		if p.curToken.Type == token.COMMA { p.nextToken() }
	}
	if p.curToken.Type == token.RPAREN { p.nextToken() }
	return node
}

func (p *Parser) parseArray() *ast.ArrayLiteral {
	array := &ast.ArrayLiteral{Elements: []ast.Node{}}
	p.nextToken()
	for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
		if p.curToken.Type != token.COMMA {
			if val := p.parseExpression(); val != nil { array.Elements = append(array.Elements, val) } else { p.nextToken() }
		}
		if p.curToken.Type == token.COMMA { p.nextToken() }
	}
	if p.curToken.Type == token.RBRACKET { p.nextToken() }
	return array
}

func (p *Parser) parseDictionary() *ast.DictionaryLiteral {
	dict := &ast.DictionaryLiteral{Elements: []ast.DictElement{}}
	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.STRING {
			key := p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == token.COLON { p.nextToken() }
			if val := p.parseExpression(); val != nil { dict.Elements = append(dict.Elements, ast.DictElement{Key: key, Value: val}) } else { p.nextToken() }
		} else if p.curToken.Type != token.COMMA { p.nextToken() }
		if p.curToken.Type == token.COMMA { p.nextToken() }
	}
	if p.curToken.Type == token.RBRACE { p.nextToken() }
	return dict
}
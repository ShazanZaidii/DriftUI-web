package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"drift/lexer"
	"drift/parser"
	"drift/ast"
	"drift/token"
)

var documentCache = make(map[string]string)

type Message struct {
	RPC    string          `json:"jsonrpc"`
	ID     *int            `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type InitializeResult struct {
	Capabilities map[string]interface{} `json:"capabilities"`
}

type TextDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Notification struct {
	RPC    string      `json:"jsonrpc"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type CompletionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type CompletionItem struct {
	Label            string `json:"label"`
	Kind             int    `json:"kind"`
	Detail           string `json:"detail"`
	InsertText       string `json:"insertText"`
	InsertTextFormat int    `json:"insertTextFormat"`
	FilterText       string `json:"filterText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type InlayHintParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Range Range `json:"range"`
}

type InlayHint struct {
	Position     Position `json:"position"`
	Label        string   `json:"label"`
	Kind         int      `json:"kind,omitempty"`
	PaddingRight bool     `json:"paddingRight,omitempty"`
}

func StartServer() {
	reader := bufio.NewReader(os.Stdin)

	for {
		var contentLength int
		for {
			header, err := reader.ReadString('\n')
			if err != nil { return }
			if header == "\r\n" || header == "\n" { break }
			if strings.HasPrefix(header, "Content-Length: ") {
				lengthStr := strings.TrimSpace(strings.TrimPrefix(header, "Content-Length: "))
				contentLength, _ = strconv.Atoi(lengthStr)
			}
		}

		if contentLength == 0 { continue }
		body := make([]byte, contentLength)
		io.ReadFull(reader, body)

		var msg Message
		json.Unmarshal(body, &msg)

		func() {
			defer func() { if r := recover(); r != nil {} }()
			handleMessage(msg)
		}()
	}
}

func handleMessage(msg Message) {
	switch msg.Method {
	case "initialize":
		result := InitializeResult{
			Capabilities: map[string]interface{}{
				"textDocumentSync": 1,
				"completionProvider": map[string]interface{}{
					"resolveProvider":   false,
					"triggerCharacters": []string{".", "@", " "},
				},
				"inlayHintProvider": true,
			},
		}
		sendResponse(msg.ID, result)

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		json.Unmarshal(msg.Params, &params)
		documentCache[params.TextDocument.URI] = params.TextDocument.Text
		runDiagnostics(params.TextDocument.URI, params.TextDocument.Text)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		json.Unmarshal(msg.Params, &params)
		if len(params.ContentChanges) > 0 {
			text := params.ContentChanges[0].Text
			documentCache[params.TextDocument.URI] = text
			runDiagnostics(params.TextDocument.URI, text)
		}

	case "textDocument/inlayHint":
		var params InlayHintParams
		json.Unmarshal(msg.Params, &params)
		hints := computeInlayHints(params.TextDocument.URI)
		sendResponse(msg.ID, hints)

	case "textDocument/completion":
		var params CompletionParams
		json.Unmarshal(msg.Params, &params)

		text := documentCache[params.TextDocument.URI]
		lines := strings.Split(text, "\n")
		var completionItems []CompletionItem

		if params.Position.Line < len(lines) {
			line := lines[params.Position.Line]
			charPos := params.Position.Character
			if charPos > len(line) { charPos = len(line) }
			currentLine := line[:charPos]

			if strings.Contains(currentLine, "set ") {
				completionItems = []CompletionItem{
					{Label: "pageTitle", FilterText: "pageTitle", Kind: 14, Detail: "Set Page Title", InsertText: "pageTitle = \"$1\"", InsertTextFormat: 2},
				}
			} else if strings.Contains(currentLine, "Color.") {
				completionItems = []CompletionItem{
					{Label: "White", Kind: 12, InsertText: "White"}, {Label: "Black", Kind: 12, InsertText: "Black"},
					{Label: "gray", Kind: 12, InsertText: "gray"}, {Label: "LightGray", Kind: 12, InsertText: "LightGray"},
					{Label: "Transparent", Kind: 12, InsertText: "Transparent"}, {Label: "Red", Kind: 12, InsertText: "Red"},
					{Label: "Green", Kind: 12, InsertText: "Green"}, {Label: "Blue", Kind: 12, InsertText: "Blue"},
				}
			} else if strings.Contains(currentLine, "ToastType.") {
				completionItems = []CompletionItem{
					{Label: "Success", Kind: 12, InsertText: "Success"},
					{Label: "Error", Kind: 12, InsertText: "Error"},
					{Label: "Info", Kind: 12, InsertText: "Info"},
				}
			} else if strings.Contains(currentLine, "ToastLength.") {
				completionItems = []CompletionItem{
					{Label: "Short", Kind: 12, InsertText: "Short"},
					{Label: "Long", Kind: 12, InsertText: "Long"},
				}
			} else if strings.Contains(currentLine, "nav.") {
				completionItems = []CompletionItem{
					{Label: "push", Kind: 2, InsertText: "push(tag = \"$1\") { $0 }", InsertTextFormat: 2},
					{Label: "pop", Kind: 2, InsertText: "pop()"},
					{Label: "replace", Kind: 2, InsertText: "replace(tag = \"$1\") { $0 }", InsertTextFormat: 2},
					{Label: "replaceRoot", Kind: 2, InsertText: "replaceRoot(tag = \"$1\") { $0 }", InsertTextFormat: 2},
				}
			} else if strings.Contains(currentLine, "Modifier.") {
				completionItems = []CompletionItem{
					{Label: "padding", FilterText: "padding", Kind: 3, Detail: "Spacing", InsertText: "padding(${1:16})", InsertTextFormat: 2},
					{Label: "fillMaxWidth", FilterText: "fillMaxWidth", Kind: 3, Detail: "Sizing", InsertText: "fillMaxWidth()", InsertTextFormat: 2},
					{Label: "fillMaxHeight", FilterText: "fillMaxHeight", Kind: 3, Detail: "Sizing", InsertText: "fillMaxHeight()", InsertTextFormat: 2},
					{Label: "fillMaxSize", FilterText: "fillMaxSize", Kind: 3, Detail: "Sizing", InsertText: "fillMaxSize()", InsertTextFormat: 2},
					{Label: "width", FilterText: "width", Kind: 3, Detail: "Sizing", InsertText: "width(${1:100})", InsertTextFormat: 2},
					{Label: "height", FilterText: "height", Kind: 3, Detail: "Sizing", InsertText: "height(${1:100})", InsertTextFormat: 2},
					{Label: "background", FilterText: "background", Kind: 3, Detail: "Color", InsertText: "background(Color.${1:White})", InsertTextFormat: 2},
				}
			} else {
				completionItems = []CompletionItem{
					{Label: "App", FilterText: "App", Kind: 7, Detail: "Root Entry Point", InsertText: "App {\n\t${1:Home}()\n}", InsertTextFormat: 2},
					{Label: "Column", FilterText: "Column", Kind: 7, Detail: "Vertical Layout", InsertText: "Column {\n\t$0\n}", InsertTextFormat: 2},
					{Label: "Row", FilterText: "Row", Kind: 7, Detail: "Horizontal Layout", InsertText: "Row {\n\t$0\n}", InsertTextFormat: 2},
					{Label: "Box", FilterText: "Box", Kind: 7, Detail: "Stack Layout", InsertText: "Box {\n\t$0\n}", InsertTextFormat: 2},
					{Label: "Text", FilterText: "Text", Kind: 7, Detail: "Text Node", InsertText: "Text(\"$1\")", InsertTextFormat: 2},
					{Label: "TextField", FilterText: "TextField", Kind: 7, Detail: "Input", InsertText: "TextField(\n\tvalue = ${1:state},\n\tonValueChange = { ${2:setState}(it) }\n)", InsertTextFormat: 2},
					{Label: "Button", FilterText: "Button", Kind: 7, Detail: "Interactive Button", InsertText: "Button(onClick = { $1 }) {\n\t$0\n}", InsertTextFormat: 2},
					{Label: "Toast", FilterText: "Toast", Kind: 7, Detail: "Popup Message", InsertText: "Toast(message = \"$1\", duration = ToastLength.Short, type = ToastType.Info)", InsertTextFormat: 2},
					{Label: "Modifier", FilterText: "Modifier", Kind: 7, Detail: "Styling Chain", InsertText: "Modifier.$0", InsertTextFormat: 2},
					{Label: "Color", FilterText: "Color", Kind: 7, Detail: "Colors", InsertText: "Color.$0", InsertTextFormat: 2},
					{Label: "@State var", FilterText: "@State", Kind: 14, Detail: "Reactive Variable", InsertText: "@State var ${1:name} = ${2:\"\"}", InsertTextFormat: 2},
					{Label: "fun", FilterText: "fun", Kind: 14, Detail: "Component Function", InsertText: "fun ${1:Name}() {\n\t$0\n}", InsertTextFormat: 2},
				}
			}
		}

		sendResponse(msg.ID, CompletionList{IsIncomplete: false, Items: completionItems})
	}
}

// AST Traversal for native IDE signature hints
func computeInlayHints(uri string) []InlayHint {
	text := documentCache[uri]
	l := lexer.New(text)
	p := parser.New(l)
	fileAST := p.ParseFile()

	var hints []InlayHint
	var walk func(node ast.Node)
	
	walk = func(node ast.Node) {
		if node == nil { return }
		switch n := node.(type) {
		case *ast.File:
			for _, c := range n.Components { walk(c) }
		case *ast.Component:
			for _, d := range n.Declarations { walk(d) }
			walk(n.UI)
		case *ast.StateDeclaration: walk(n.Value)
		case *ast.VariableDeclaration: walk(n.Value)
		case *ast.PropertyAccess: walk(n.ObjectNode)
		case *ast.MethodCall:
			walk(n.CallerNode)
			checkArgs(n.Arguments, getMethodSig(n.Method), &hints)
			walk(n.Block)
		case *ast.FunctionCall:
			walk(n.CallerNode)
			if id, ok := n.CallerNode.(*ast.Identifier); ok {
				checkArgs(n.Arguments, getFuncSig(id.Value), &hints)
			}
			walk(n.Block)
		case *ast.IfExpression:
			walk(n.Condition); walk(n.Block); walk(n.Alternative)
		case *ast.InfixExpression:
			walk(n.Left); walk(n.Right)
		case *ast.Assignment: walk(n.Value)
		case *ast.BlockLiteral:
			for _, s := range n.Statements { walk(s) }
		case *ast.ComponentCall:
			checkArgs(n.Arguments, getComponentSig(n.Name), &hints)
			for _, c := range n.Children { walk(c) }
		case *ast.Column: for _, c := range n.Children { walk(c) }
		case *ast.Row: for _, c := range n.Children { walk(c) }
		case *ast.Box: for _, c := range n.Children { walk(c) }
		case *ast.Text: walk(n.Expression)
		}
	}
	
	if fileAST != nil { walk(fileAST) }
	return hints
}

func checkArgs(args []ast.Argument, sig []string, hints *[]InlayHint) {
	if len(sig) == 0 { return }
	for i, arg := range args {
		if !arg.IsNamed && i < len(sig) {
			*hints = append(*hints, InlayHint{
				Position: Position{Line: arg.Line - 1, Character: arg.Column - 1},
				Label:    sig[i] + ":",
				Kind:     2,
				PaddingRight: true,
			})
		}
	}
}

func getFuncSig(name string) []string {
	switch name {
	case "Toast": return []string{"message", "duration", "type"}
	}
	return nil
}

func getComponentSig(name string) []string {
	switch name {
	case "TextField": return []string{"value", "onValueChange", "modifier", "singleLine", "placeholder"}
	case "Button": return []string{"onClick", "modifier"}
	case "Text": return []string{"text", "modifier"}
	case "Toast": return []string{"message", "duration", "type"}
	}
	return nil
}

func getMethodSig(name string) []string {
	switch name {
	case "push", "replace", "replaceRoot": return []string{"tag"}
	}
	return nil
}

func runDiagnostics(uri string, text string) {
	l := lexer.New(text)
	p := parser.New(l)
	fileAST := p.ParseFile()

	diagnostics := []Diagnostic{}

	for _, err := range p.Errors {
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: err.Line - 1, Character: err.Column - 1},
				End:   Position{Line: err.Line - 1, Character: err.Column + 4},
			},
			Severity: 1,
			Source:   "Drift Syntax",
			Message:  err.Message,
		})
	}

	if fileAST != nil {
		defined := map[string]bool{
			"Column": true, "Row": true, "Box": true, "Text": true,
			"Modifier": true, "State": true, "App": true, "Button": true,
			"TextField": true, "Color": true, "Toast": true, "ToastType": true, "ToastLength": true,
			"TextStyle": true, "TextFieldDefaults": true, "scaledSp": true, "scaledDp": true,
			"PasswordVisualTransformation": true,
		}

		validModifiers := map[string]bool{
			"padding": true, "fillMaxWidth": true, "background": true,
			"fillMaxHeight": true, "fillMaxSize": true, "width": true,
			"height": true, "size": true, "weight": true, "align": true,
			"cornerRadius": true, "clip": true, "border": true, "alpha": true,
			"shadow": true, "clickable": true, "offset": true, "type": true,
			"dp": true, "sp": true, "f": true,
			"color": true, "fontSize": true, "fontWeight": true,
		}

		for _, comp := range fileAST.Components { defined[comp.Name] = true }

		l2 := lexer.New(text)
		inModifierChain := false
		var prevToken token.Token

		for tok := l2.NextToken(); tok.Type != token.EOF; tok = l2.NextToken() {
			if tok.Type == token.IDENT && len(tok.Literal) > 0 && tok.Literal[0] >= 'A' && tok.Literal[0] <= 'Z' {
				if prevToken.Type != token.DOT && !defined[tok.Literal] {
					diagnostics = append(diagnostics, Diagnostic{
						Range: Range{
							Start: Position{Line: tok.Line - 1, Character: tok.Column - 1},
							End:   Position{Line: tok.Line - 1, Character: tok.Column - 1 + len(tok.Literal)},
						},
						Severity: 1,
						Source:   "Drift Semantic",
						Message:  fmt.Sprintf("Unknown component or reference: '%s'", tok.Literal),
					})
				}
			}

			if tok.Literal == "Modifier" { inModifierChain = true }

			if inModifierChain && tok.Type == token.IDENT && tok.Literal != "Modifier" {
				if prevToken.Type == token.DOT {
					if !validModifiers[tok.Literal] {
						diagnostics = append(diagnostics, Diagnostic{
							Range: Range{
								Start: Position{Line: tok.Line - 1, Character: tok.Column - 1},
								End:   Position{Line: tok.Line - 1, Character: tok.Column - 1 + len(tok.Literal)},
							},
							Severity: 1,
							Source:   "Drift Styling",
							Message:  fmt.Sprintf("Invalid Modifier: '%s'", tok.Literal),
						})
					}
				} else {
					inModifierChain = false
				}
			}

			if tok.Type != token.IDENT && tok.Type != token.DOT && tok.Type != token.LPAREN && tok.Type != token.RPAREN && tok.Type != token.INT && tok.Type != token.STRING && tok.Type != token.COMMA {
				inModifierChain = false
			}

			prevToken = tok
		}
	}

	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func sendResponse(id *int, result interface{}) {
	if id == nil { return }
	send(map[string]interface{}{"jsonrpc": "2.0", "id": *id, "result": result})
}

func sendNotification(method string, params interface{}) {
	send(Notification{RPC: "2.0", Method: method, Params: params})
}

func send(data interface{}) {
	body, _ := json.Marshal(data)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	os.Stdout.Write([]byte(header))
	os.Stdout.Write(body)
}
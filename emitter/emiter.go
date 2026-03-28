package emitter

import (
	"drift/ast"
	"fmt"
	"regexp"
	"strings"
)

type Emitter struct{}

func New() *Emitter { return &Emitter{} }

func extractClickable(mod *ast.ModifierChain, e *Emitter) string {
	if mod == nil { return "" }
	for _, call := range mod.Calls {
		if call.Name == "clickable" && len(call.Arguments) > 0 {
			return fmt.Sprintf(` onClick={() => { %s }}`, strings.TrimSpace(e.Emit(call.Arguments[0])))
		}
	}
	return ""
}

func resolveArgsMap(args []ast.Argument, sig []string) map[string]ast.Node {
	m := make(map[string]ast.Node)
	for i, arg := range args {
		if arg.IsNamed {
			m[arg.Name] = arg.Value
		} else if i < len(sig) {
			m[sig[i]] = arg.Value
		}
	}
	return m
}

func (e *Emitter) emitModifier(mod *ast.ModifierChain, baseStyle string) string {
	styleMap := make(map[string]string)

	if baseStyle != "" {
		pairs := strings.Split(baseStyle, ";")
		for _, p := range pairs {
			kv := strings.SplitN(p, ":", 2)
			if len(kv) == 2 {
				styleMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	if mod != nil {
		for _, call := range mod.Calls {
			switch call.Name {
			case "padding":
				if len(call.Arguments) == 1 { styleMap["padding"] = fmt.Sprintf("'%spx'", e.Emit(call.Arguments[0])) }
				if len(call.Arguments) == 4 { styleMap["padding"] = fmt.Sprintf("'%spx %spx %spx %spx'", e.Emit(call.Arguments[0]), e.Emit(call.Arguments[1]), e.Emit(call.Arguments[2]), e.Emit(call.Arguments[3])) }
			case "offset": if len(call.Arguments) > 0 { styleMap["transform"] = fmt.Sprintf("`translate(${%s}px, 0px)`", e.Emit(call.Arguments[0])) }
			case "fillMaxWidth": styleMap["width"] = "'100%'"
			case "fillMaxHeight": styleMap["height"] = "'100%'"
			case "fillMaxSize": styleMap["width"] = "'100%'"; styleMap["height"] = "'100%'"
			case "width": if len(call.Arguments) > 0 { styleMap["width"] = fmt.Sprintf("'%spx'", e.Emit(call.Arguments[0])) }
			case "height": if len(call.Arguments) > 0 { styleMap["height"] = fmt.Sprintf("'%spx'", e.Emit(call.Arguments[0])) }
			case "size": if len(call.Arguments) > 0 { val := e.Emit(call.Arguments[0]); styleMap["width"] = fmt.Sprintf("'%spx'", val); styleMap["height"] = fmt.Sprintf("'%spx'", val) }
			case "weight": if len(call.Arguments) > 0 { styleMap["flex"] = fmt.Sprintf("'%s 1 0%%'", e.Emit(call.Arguments[0])) }
			case "align":
				if len(call.Arguments) > 0 {
					arg := e.Emit(call.Arguments[0])
					if arg == `"center"` {
						styleMap["margin"] = "'auto'"
					} else {
						styleMap["alignSelf"] = arg
					}
				}
			case "cornerRadius", "clip": if len(call.Arguments) > 0 { styleMap["borderRadius"] = fmt.Sprintf("'%spx'", e.Emit(call.Arguments[0])) }
			case "border": if len(call.Arguments) == 2 { styleMap["border"] = fmt.Sprintf("'%spx solid ' + %s", e.Emit(call.Arguments[0]), e.Emit(call.Arguments[1])) }
			case "alpha": if len(call.Arguments) > 0 { styleMap["opacity"] = e.Emit(call.Arguments[0]) }
			case "shadow": if len(call.Arguments) > 0 { styleMap["boxShadow"] = fmt.Sprintf("'0px 4px %spx rgba(0,0,0,0.15)'", e.Emit(call.Arguments[0])) }
			case "clickable": styleMap["cursor"] = "'pointer'"
			case "background": if len(call.Arguments) > 0 { styleMap["backgroundColor"] = e.Emit(call.Arguments[0]) }
			case "color": if len(call.Arguments) > 0 { styleMap["color"] = e.Emit(call.Arguments[0]) }
			case "fontSize": if len(call.Arguments) > 0 { styleMap["fontSize"] = fmt.Sprintf("'%spx'", e.Emit(call.Arguments[0])) }
			case "fontWeight": if len(call.Arguments) > 0 { styleMap["fontWeight"] = fmt.Sprintf("%s", e.Emit(call.Arguments[0])) }
			}
		}
	}

	if len(styleMap) == 0 { return "" }

	var styles []string
	for k, v := range styleMap {
		styles = append(styles, fmt.Sprintf("%s: %s", k, v))
	}
	return fmt.Sprintf(" style={{ %s }}", strings.Join(styles, ", "))
}

func (e *Emitter) Emit(node ast.Node) string {
	switch n := node.(type) {
	case *ast.File:
		var out strings.Builder
		out.WriteString("import React, { useEffect, useState } from 'react';\n")
		out.WriteString("import { useNav } from './drift_router.jsx';\n")
		out.WriteString("import { useToast } from './drift_toast.jsx';\n\n")

		out.WriteString("const scaledSp = (val) => val;\nconst scaledDp = (val) => val;\n")
		out.WriteString("const ToastType = { Error: 'error', Success: 'success', Info: 'info' };\n")
		out.WriteString("const ToastLength = { Short: 'short', Long: 'long' };\n")
		out.WriteString("const Color = {\n  Transparent: 'transparent',\n  White: '#FFFFFF', Black: '#000000',\n  gray: '#9CA3AF', Gray: '#9CA3AF', LightGray: '#D1D5DB',\n  Red: '#EF4444', Green: '#22C55E', Blue: '#3B82F6'\n};\n")
		out.WriteString("Object.keys(Color).forEach(k => {\n  const hex = Color[k];\n  const c = new String(hex);\n  c.copy = (args) => { if (args && args.alpha !== undefined) return `rgba(128,128,128,${args.alpha})`; return hex; };\n  Color[k] = c;\n});\n\n")

		for _, comp := range n.Components { out.WriteString(e.Emit(comp)) }
		return out.String()

	case *ast.Component:
		var body strings.Builder
		if title, exists := n.Metadata["pageTitle"]; exists {
			body.WriteString(fmt.Sprintf("  useEffect(() => { document.title = \"%s\"; }, []);\n", title))
		}
		for _, decl := range n.Declarations { body.WriteString("  " + e.Emit(decl)) }
		uiOutput := ""
		if n.UI != nil { uiOutput = e.Emit(n.UI) }
		return fmt.Sprintf("export function %s() {\n  const toast = useToast();\n%s\n  return (\n%s  );\n}\n", n.Name, body.String(), uiOutput)

	case *ast.StateDeclaration:
		setter := "set" + strings.ToUpper(string(n.Name[0])) + n.Name[1:]
		return fmt.Sprintf("const [%s, %s] = useState(%s);\n", n.Name, setter, e.Emit(n.Value))

	case *ast.VariableDeclaration:
		keyword := "var"
		if n.IsVal { keyword = "const" } else { keyword = "let" }
		return fmt.Sprintf("%s %s = %s;\n", keyword, n.Name, e.Emit(n.Value))

	case *ast.FunctionCall:
		caller := e.Emit(n.CallerNode)
		if caller == "Toast" {
			argsMap := resolveArgsMap(n.Arguments, []string{"message", "duration", "type"})
			msg, duration, tType := "''", "ToastLength.Short", "ToastType.Info"
			if val, ok := argsMap["message"]; ok { msg = e.Emit(val) }
			if val, ok := argsMap["duration"]; ok { duration = e.Emit(val) }
			if val, ok := argsMap["type"]; ok { tType = e.Emit(val) }
			return fmt.Sprintf("toast.show(%s, %s, %s)", msg, duration, tType)
		}
		var args []string
		for _, arg := range n.Arguments { args = append(args, e.Emit(arg.Value)) }
		return fmt.Sprintf("%s(%s)", caller, strings.Join(args, ", "))

	case *ast.PropertyAccess:
		return fmt.Sprintf("%s.%s", e.Emit(n.ObjectNode), n.Property)

	case *ast.MethodCall:
		caller := e.Emit(n.CallerNode)
		
		if caller == "nav" && (n.Method == "push" || n.Method == "replace" || n.Method == "replaceRoot") {
			argsMap := resolveArgsMap(n.Arguments, []string{"tag"})
			tagArg := "null"
			if val, ok := argsMap["tag"]; ok { tagArg = e.Emit(val) }
			
			componentJSX := "null"
			if n.Block != nil {
				if block, ok := n.Block.(*ast.BlockLiteral); ok && len(block.Statements) > 0 {
					componentJSX = strings.TrimSpace(e.Emit(block.Statements[len(block.Statements)-1]))
					componentJSX = strings.TrimSuffix(componentJSX, ";")
				} else {
					componentJSX = strings.TrimSpace(e.Emit(n.Block))
					componentJSX = strings.TrimSuffix(componentJSX, ";")
				}
			}
			return fmt.Sprintf("%s.%s(%s, %s)", caller, n.Method, componentJSX, tagArg)
		} else if caller == "nav" && n.Method == "pop" {
			return fmt.Sprintf("%s.pop()", caller)
		}

		var args []string
		for _, arg := range n.Arguments { args = append(args, e.Emit(arg.Value)) }
		return fmt.Sprintf("%s.%s(%s)", caller, n.Method, strings.Join(args, ", "))

	case *ast.IfExpression:
		res := fmt.Sprintf("if (%s) {\n %s \n}", e.Emit(n.Condition), e.Emit(n.Block))
		if n.Alternative != nil { res += fmt.Sprintf(" else {\n %s \n}", e.Emit(n.Alternative)) }
		return res

	case *ast.Assignment:
		setter := "set" + strings.ToUpper(string(n.Name[0])) + n.Name[1:]
		return fmt.Sprintf("%s(%s)", setter, e.Emit(n.Value))

	case *ast.InfixExpression:
		op := n.Operator
		if op == "==" { op = "===" }
		if op == "!=" { op = "!==" }
		return fmt.Sprintf("%s %s %s", e.Emit(n.Left), op, e.Emit(n.Right))

	case *ast.BlockLiteral:
		var out strings.Builder
		for _, s := range n.Statements {
			stmt := strings.TrimSpace(e.Emit(s))
			if !strings.HasSuffix(stmt, ";") && !strings.HasPrefix(stmt, "if") && !strings.HasPrefix(stmt, "toast") { stmt += ";" }
			out.WriteString("        " + stmt + "\n")
		}
		return out.String()

	case *ast.Identifier: return n.Value
	case *ast.StringLiteral:
		re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
		return fmt.Sprintf("`%s`", re.ReplaceAllString(n.Value, "$${$1}"))
	case *ast.IntLiteral:
		val := strings.TrimSuffix(n.Value, "f")
		val = strings.TrimSuffix(val, "F")
		return val
	case *ast.BooleanLiteral: if n.Value { return "true" } else { return "false" }
	case *ast.ArrayLiteral:
		var elements []string
		for _, el := range n.Elements { elements = append(elements, e.Emit(el)) }
		return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
	case *ast.DictionaryLiteral:
		var elements []string
		for _, el := range n.Elements { elements = append(elements, fmt.Sprintf("\"%s\": %s", el.Key, e.Emit(el.Value))) }
		return fmt.Sprintf("{ %s }", strings.Join(elements, ", "))

	case *ast.ComponentCall:
		if n.Name == "Toast" {
			argsMap := resolveArgsMap(n.Arguments, []string{"message", "duration", "type"})
			msg, duration, tType := "''", "ToastLength.Short", "ToastType.Info"
			if val, ok := argsMap["message"]; ok { msg = e.Emit(val) }
			if val, ok := argsMap["duration"]; ok { duration = e.Emit(val) }
			if val, ok := argsMap["type"]; ok { tType = e.Emit(val) }
			return fmt.Sprintf("      toast.show(%s, %s, %s);\n", msg, duration, tType)
		}
		if n.Name == "Button" {
			onClickAttr := extractClickable(n.Modifier, e)
			argsMap := resolveArgsMap(n.Arguments, []string{"onClick", "modifier"})
			
			if val, ok := argsMap["onClick"]; ok {
				onClickAttr = fmt.Sprintf(` onClick={() => { %s }}`, strings.TrimSpace(e.Emit(val)))
			}
			var children strings.Builder
			for _, child := range n.Children { children.WriteString(e.Emit(child)) }
			baseStyle := `backgroundColor: '#1E1E1E'; color: 'white'; padding: '12px 24px'; borderRadius: '8px'; border: 'none'; cursor: 'pointer'; fontWeight: '600'; display: 'flex'; alignItems: 'center'; justifyContent: 'center'; width: 'fit-content'`
			return fmt.Sprintf("      <button%s%s>\n%s      </button>\n", e.emitModifier(n.Modifier, baseStyle), onClickAttr, children.String())
		}
		if n.Name == "TextField" {
			var props strings.Builder
			var inlineStyles = "padding: '14px'; borderRadius: '8px'; border: '1px solid #E0E0E0'; outline: 'none'; boxSizing: 'border-box'; fontFamily: 'system-ui, sans-serif'"
			isTextArea := false

			argsMap := resolveArgsMap(n.Arguments, []string{"value", "onValueChange", "modifier", "singleLine", "placeholder", "textStyle", "colors", "visualTransformation"})

			if val, ok := argsMap["value"]; ok { props.WriteString(fmt.Sprintf(` value={%s}`, strings.TrimSpace(e.Emit(val)))) }
			if val, ok := argsMap["onValueChange"]; ok {
				lambdaBody := strings.TrimSpace(e.Emit(val))
				props.WriteString(fmt.Sprintf(` onChange={(e) => { 
					const rawVal = e.target.value;
					const it = new String(rawVal);
					it.text = rawVal;
					it.copy = (args) => (args && args.text !== undefined) ? args.text : rawVal;
					it.take = (n) => rawVal.substring(0, n);
					%s 
				}}`, lambdaBody))
			}
			if val, ok := argsMap["singleLine"]; ok && e.Emit(val) == "false" { isTextArea = true }
			if val, ok := argsMap["placeholder"]; ok {
				if block, ok := val.(*ast.BlockLiteral); ok && len(block.Statements) > 0 {
					if comp, ok := block.Statements[0].(*ast.ComponentCall); ok && comp.Name == "Text" && len(comp.Arguments) > 0 {
						if strLit, ok := comp.Arguments[0].Value.(*ast.StringLiteral); ok {
							props.WriteString(fmt.Sprintf(` placeholder=%s`, e.Emit(strLit)))
						}
					}
				}
			}
			if val, ok := argsMap["textStyle"]; ok {
				if fc, ok := val.(*ast.FunctionCall); ok {
					for _, tsArg := range fc.Arguments {
						if tsArg.Name == "fontSize" { inlineStyles += fmt.Sprintf("; fontSize: %s + 'px'", e.Emit(tsArg.Value)) }
						if tsArg.Name == "color" { inlineStyles += fmt.Sprintf("; color: %s", e.Emit(tsArg.Value)) }
					}
				}
			}
			if val, ok := argsMap["colors"]; ok {
				if mc, ok := val.(*ast.MethodCall); ok {
					for _, cArg := range mc.Arguments {
						if cArg.Name == "focusedContainerColor" || cArg.Name == "unfocusedContainerColor" { inlineStyles += fmt.Sprintf("; backgroundColor: %s", e.Emit(cArg.Value)) }
						if cArg.Name == "focusedIndicatorColor" || cArg.Name == "unfocusedIndicatorColor" { inlineStyles += fmt.Sprintf("; borderColor: %s", e.Emit(cArg.Value)) }
					}
				}
			}
			if val, ok := argsMap["visualTransformation"]; ok && strings.Contains(e.Emit(val), "PasswordVisualTransformation") { props.WriteString(` type="password"`) }

			style := e.emitModifier(n.Modifier, inlineStyles)
			if isTextArea { return fmt.Sprintf("      <textarea%s%s />\n", style, props.String()) }
			return fmt.Sprintf("      <input%s%s />\n", style, props.String())
		}

		var props strings.Builder
		onClickAttr := extractClickable(n.Modifier, e)
		if n.Modifier != nil {
			if styleStr := e.emitModifier(n.Modifier, ""); styleStr != "" { props.WriteString(styleStr) }
		}
		for _, arg := range n.Arguments { props.WriteString(fmt.Sprintf(" %s={%s}", arg.Name, strings.TrimSpace(e.Emit(arg.Value)))) }
		if len(n.Children) > 0 {
			var children strings.Builder
			for _, child := range n.Children { children.WriteString(e.Emit(child)) }
			return fmt.Sprintf("      <%s%s%s>\n%s      </%s>\n", n.Name, props.String(), onClickAttr, children.String(), n.Name)
		}
		return fmt.Sprintf("      <%s%s%s />\n", n.Name, props.String(), onClickAttr)

	case *ast.Column:
		var out strings.Builder
		style := e.emitModifier(n.Modifier, "display: 'flex'; flexDirection: 'column'; boxSizing: 'border-box'")
		out.WriteString(fmt.Sprintf("    <div%s%s>\n", style, extractClickable(n.Modifier, e)))
		for _, child := range n.Children { out.WriteString(e.Emit(child)) }
		out.WriteString("    </div>\n")
		return out.String()

	case *ast.Row:
		var out strings.Builder
		style := e.emitModifier(n.Modifier, "display: 'flex'; flexDirection: 'row'; boxSizing: 'border-box'")
		out.WriteString(fmt.Sprintf("    <div%s%s>\n", style, extractClickable(n.Modifier, e)))
		for _, child := range n.Children { out.WriteString(e.Emit(child)) }
		out.WriteString("    </div>\n")
		return out.String()

	case *ast.Box:
		var out strings.Builder
		style := e.emitModifier(n.Modifier, "display: 'grid'; boxSizing: 'border-box'")
		out.WriteString(fmt.Sprintf("    <div%s%s>\n", style, extractClickable(n.Modifier, e)))
		for _, child := range n.Children {
			// FIX: Forced width: 100% and height: 100% so margin: auto has cross-axis boundaries to center against.
			out.WriteString(fmt.Sprintf("      <div style={{ gridArea: '1 / 1', display: 'flex', width: '100%%', height: '100%%' }}>\n%s      </div>\n", e.Emit(child)))
		}
		out.WriteString("    </div>\n")
		return out.String()

	case *ast.Text:
		style := e.emitModifier(n.Modifier, "fontFamily: 'system-ui, sans-serif'; margin: 0")
		return fmt.Sprintf("      <span%s%s>{%s}</span>\n", style, extractClickable(n.Modifier, e), e.Emit(n.Expression))
	}
	return ""
}
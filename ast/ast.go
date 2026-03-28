package ast

type Node interface {
	IsNode()
}

type File struct {
	AppRoot    *AppRoot
	Components []*Component
}

type AppRoot struct {
	EntryComponent string
}

type Component struct {
	Name         string
	Metadata     map[string]string
	Declarations []Node
	UI           Node
}

type StateDeclaration struct {
	Name  string
	Value Node
}

type VariableDeclaration struct {
	IsVal bool
	Name  string
	Value Node
}

type StringLiteral struct{ Value string }
type IntLiteral struct{ Value string }
type BooleanLiteral struct{ Value bool }
type ArrayLiteral struct{ Elements []Node }
type DictionaryLiteral struct{ Elements []DictElement }

type DictElement struct {
	Key   string
	Value Node
}

type Identifier struct {
	Value string
}

type PropertyAccess struct {
	ObjectNode Node
	Property   string
}

type MethodCall struct {
	CallerNode Node
	Method     string
	Arguments  []Argument
	Block      Node
}

type FunctionCall struct {
	CallerNode Node
	Arguments  []Argument
	Block      Node
}

type IfExpression struct {
	Condition   Node
	Block       Node
	Alternative Node
}

type InfixExpression struct {
	Left     Node
	Operator string
	Right    Node
}

type Assignment struct {
	Name  string
	Value Node
}

type BlockLiteral struct {
	Statements []Node
}

// Upgraded to track precise location and dev-intent (Named vs Positional)
type Argument struct {
	Name    string
	Value   Node
	IsNamed bool
	Line    int
	Column  int
}

type ModifierChain struct {
	Calls []ModifierCall
}

type ModifierCall struct {
	Name      string
	Arguments []Node
}

type ComponentCall struct {
	Name      string
	Arguments []Argument
	Modifier  *ModifierChain
	Children  []Node
}

type Column struct {
	Modifier *ModifierChain
	Children []Node
}
type Row struct {
	Modifier *ModifierChain
	Children []Node
}
type Box struct {
	Modifier *ModifierChain
	Children []Node
}
type Text struct {
	Expression Node
	Modifier   *ModifierChain
}

func (f *File) IsNode()                {}
func (a *AppRoot) IsNode()             {}
func (c *Component) IsNode()           {}
func (s *StateDeclaration) IsNode()    {}
func (v *VariableDeclaration) IsNode() {}
func (s *StringLiteral) IsNode()       {}
func (i *IntLiteral) IsNode()          {}
func (b *BooleanLiteral) IsNode()      {}
func (a *ArrayLiteral) IsNode()        {}
func (d *DictionaryLiteral) IsNode()   {}
func (i *Identifier) IsNode()          {}
func (p *PropertyAccess) IsNode()      {}
func (m *MethodCall) IsNode()          {}
func (f *FunctionCall) IsNode()        {}
func (i *IfExpression) IsNode()        {}
func (i *InfixExpression) IsNode()     {}
func (a *Assignment) IsNode()          {}
func (b *BlockLiteral) IsNode()        {}
func (c *ComponentCall) IsNode()       {}
func (m *ModifierChain) IsNode()       {}
func (c *Column) IsNode()              {}
func (r *Row) IsNode()                 {}
func (b *Box) IsNode()                 {}
func (t *Text) IsNode()                {}
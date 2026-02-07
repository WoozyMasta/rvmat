package rvmat

// valueKind represents the kind of a parsed value.
type valueKind int

const (
	// valueNumber indicates numeric literal.
	valueNumber valueKind = iota
	// valueString indicates quoted string literal.
	valueString
	// valueIdent indicates bare identifier literal.
	valueIdent
	// valueArray indicates array literal.
	valueArray
)

// value represents a parsed value.
type value struct {
	Str   string    // String value
	Array []value   // Array value
	Kind  valueKind // Value kind
	Num   float64   // Number value
}

// node is a parsed AST node.
type node interface {
	node()
}

// assignNode represents name[ ] = value; assignments.
type assignNode struct {
	Name    string // Name of the assigned variable
	Value   value  // Value of the assignment
	IsArray bool   // Whether the assignment is an array
}

// node implements the Node interface.
func (assignNode) node() {}

// classNode represents class blocks.
type classNode struct {
	Name string // Name of the class
	Base string // Base class name
	Body []node // Body of the class
}

// node implements the Node interface.
func (classNode) node() {}

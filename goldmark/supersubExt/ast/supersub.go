// Package ast defines AST nodes that represents extension's elements
package ast

import (
	gast "goDemo/gmark/goldmark/ast"
)

// A Subscript struct represents subscript text.
type SuperSubScript struct {
	gast.BaseInline
	Super bool
}

// Dump implements Node.Dump.
func (n *SuperSubScript) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindSubscript is a NodeKind of the Subscript node.
var KindSuperSubScript = gast.NewNodeKind("SuperSubScript")

// Kind implements Node.Kind.
func (n *SuperSubScript) Kind() gast.NodeKind {
	return KindSuperSubScript
}

// NewSubscript returns a new Subscript node.
func NewSuperSubScript() *SuperSubScript {
	return &SuperSubScript{}
}

package ast

import (
	"biscuit/compiler/ast/nodes"
)

// Stmt is the interface that all statement nodes implement
type Stmt interface {
	nodes.Stmt
}

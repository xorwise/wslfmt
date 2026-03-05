package fixer

import (
	"go/token"

	"github.com/dave/dst"
)

// isAssignLike returns true if stmt is an assignment (=/,:=) or increment/decrement.
func isAssignLike(stmt dst.Stmt) bool {
	switch stmt.(type) {
	case *dst.AssignStmt, *dst.IncDecStmt:
		return true
	}
	return false
}

// isAssignment returns true if stmt is an assignment (= or :=).
func isAssignment(stmt dst.Stmt) bool {
	_, ok := stmt.(*dst.AssignStmt)
	return ok
}

// isShortDecl returns true if stmt is a short variable declaration (:=).
func isShortDecl(stmt dst.Stmt) bool {
	a, ok := stmt.(*dst.AssignStmt)
	return ok && a.Tok == token.DEFINE
}

// isBlockStmt returns true if stmt contains a body block (if/for/switch/select/typeSwitch).
func isBlockStmt(stmt dst.Stmt) bool {
	switch stmt.(type) {
	case *dst.IfStmt, *dst.ForStmt, *dst.RangeStmt,
		*dst.SwitchStmt, *dst.TypeSwitchStmt, *dst.SelectStmt:
		return true
	}
	return false
}

// isBranch returns true if stmt is return/break/continue/goto/fallthrough.
func isBranch(stmt dst.Stmt) bool {
	switch stmt.(type) {
	case *dst.ReturnStmt, *dst.BranchStmt:
		return true
	}
	return false
}

// extractAssignedIdents returns the names of variables assigned in an AssignStmt.
func extractAssignedIdents(stmt dst.Stmt) []string {
	a, ok := stmt.(*dst.AssignStmt)
	if !ok {
		return nil
	}
	var names []string
	for _, lhs := range a.Lhs {
		if ident, ok := lhs.(*dst.Ident); ok && ident.Name != "_" {
			names = append(names, ident.Name)
		}
	}
	return names
}

// extractDeclaredIdents returns variable names declared in a var/const DeclStmt.
func extractDeclaredIdents(stmt dst.Stmt) []string {
	decl, ok := stmt.(*dst.DeclStmt)
	if !ok {
		return nil
	}
	gen, ok := decl.Decl.(*dst.GenDecl)
	if !ok {
		return nil
	}
	var names []string
	for _, spec := range gen.Specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if name.Name != "_" {
				names = append(names, name.Name)
			}
		}
	}
	return names
}

// extractAllVars returns variables assigned, declared, or inc/dec'd in stmt.
func extractAllVars(stmt dst.Stmt) []string {
	if names := extractAssignedIdents(stmt); len(names) > 0 {
		return names
	}
	if names := extractDeclaredIdents(stmt); len(names) > 0 {
		return names
	}
	if inc, ok := stmt.(*dst.IncDecStmt); ok {
		if ident, ok := inc.X.(*dst.Ident); ok && ident.Name != "_" {
			return []string{ident.Name}
		}
	}
	return nil
}

// usesAnyIdent returns true if stmt references any of the given identifier names.
func usesAnyIdent(stmt dst.Stmt, names []string) bool {
	if len(names) == 0 {
		return false
	}
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	found := false
	dst.Inspect(stmt, func(node dst.Node) bool {
		if found {
			return false
		}
		if ident, ok := node.(*dst.Ident); ok {
			if nameSet[ident.Name] {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// usesAnyIdentInHeader returns true if any of the given names appear in the
// "header" of a control-flow statement (condition, init, post, tag — not body).
// For non-control-flow statements, falls back to usesAnyIdent.
func usesAnyIdentInHeader(stmt dst.Stmt, names []string) bool {
	if len(names) == 0 {
		return false
	}
	switch s := stmt.(type) {
	case *dst.IfStmt:
		var nodes []dst.Node
		if s.Init != nil {
			nodes = append(nodes, s.Init)
		}
		nodes = append(nodes, s.Cond)
		return anyNodeUsesIdents(nodes, names)
	case *dst.ForStmt:
		var nodes []dst.Node
		if s.Init != nil {
			nodes = append(nodes, s.Init)
		}
		if s.Cond != nil {
			nodes = append(nodes, s.Cond)
		}
		if s.Post != nil {
			nodes = append(nodes, s.Post)
		}
		return anyNodeUsesIdents(nodes, names)
	case *dst.RangeStmt:
		var nodes []dst.Node
		if s.Key != nil {
			nodes = append(nodes, s.Key)
		}
		if s.Value != nil {
			nodes = append(nodes, s.Value)
		}
		nodes = append(nodes, s.X)
		return anyNodeUsesIdents(nodes, names)
	case *dst.SwitchStmt:
		var nodes []dst.Node
		if s.Init != nil {
			nodes = append(nodes, s.Init)
		}
		if s.Tag != nil {
			nodes = append(nodes, s.Tag)
		}
		return anyNodeUsesIdents(nodes, names)
	case *dst.TypeSwitchStmt:
		var nodes []dst.Node
		if s.Init != nil {
			nodes = append(nodes, s.Init)
		}
		nodes = append(nodes, s.Assign)
		return anyNodeUsesIdents(nodes, names)
	}
	return usesAnyIdent(stmt, names)
}

// anyNodeUsesIdents returns true if any of the given names appear in any of the nodes.
func anyNodeUsesIdents(nodes []dst.Node, names []string) bool {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	found := false
	for _, node := range nodes {
		dst.Inspect(node, func(n dst.Node) bool {
			if found {
				return false
			}
			if ident, ok := n.(*dst.Ident); ok && nameSet[ident.Name] {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

// isErrorCheckIf returns true if stmt is an `if err != nil` or similar error check.
func isErrorCheckIf(stmt dst.Stmt) bool {
	ifStmt, ok := stmt.(*dst.IfStmt)
	if !ok {
		return false
	}
	bin, ok := ifStmt.Cond.(*dst.BinaryExpr)
	if !ok {
		return false
	}
	if bin.Op != token.NEQ && bin.Op != token.EQL {
		return false
	}
	return referencesErr(bin.X) || referencesErr(bin.Y)
}

func referencesErr(expr dst.Expr) bool {
	ident, ok := expr.(*dst.Ident)
	return ok && ident.Name == "err"
}

// appendedIdent returns the name of the variable being appended to if stmt is
// of the form `x = append(x, ...)` or `x := append(x, ...)`, otherwise "".
func appendedIdent(stmt dst.Stmt) string {
	a, ok := stmt.(*dst.AssignStmt)
	if !ok || len(a.Rhs) != 1 {
		return ""
	}
	call, ok := a.Rhs[0].(*dst.CallExpr)
	if !ok {
		return ""
	}
	fun, ok := call.Fun.(*dst.Ident)
	if !ok || fun.Name != "append" {
		return ""
	}
	if len(call.Args) == 0 {
		return ""
	}
	ident, ok := call.Args[0].(*dst.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

// prevHasVar returns true if prev assigns, declares, or inc/dec's the given name,
// or uses it as an expression.
func prevHasVar(prev dst.Stmt, name string) bool {
	for _, n := range extractAllVars(prev) {
		if n == name {
			return true
		}
	}
	return usesAnyIdent(prev, []string{name})
}

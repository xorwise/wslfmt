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

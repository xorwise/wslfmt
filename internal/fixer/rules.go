package fixer

import (
	"github.com/dave/dst"
)

// shouldHaveBlankLine returns true if a blank line should be inserted between prev and curr.
// blockLen is the total number of statements in the containing block.
func shouldHaveBlankLine(prev, curr dst.Stmt, blockLen int) bool {
	// branch/return: not cuddled if block has more than 2 lines.
	if isBranch(curr) && blockLen > 2 {
		return true
	}

	// label: never cuddled.
	if _, ok := curr.(*dst.LabeledStmt); ok {
		return true
	}

	// decl: declarations (var/const/type) should never be cuddled.
	if _, ok := curr.(*dst.DeclStmt); ok {
		return true
	}

	// append: only cuddle with statement that uses/assigns/declares the appended var.
	if appended := appendedIdent(curr); appended != "" {
		return !prevHasVar(prev, appended)
	}

	// assign and inc-dec: only cuddle with other assign or inc-dec.
	if isAssignLike(curr) {
		return !isAssignLike(prev)
	}

	// defer: cuddle with other defer, or with prev that assigns/declares a used var.
	if _, ok := curr.(*dst.DeferStmt); ok {
		if _, ok := prev.(*dst.DeferStmt); ok {
			return false
		}
		return !prevAssignsUsedVar(prev, curr)
	}

	// go: cuddle with other go, or with prev that assigns/declares a used var.
	if _, ok := curr.(*dst.GoStmt); ok {
		if _, ok := prev.(*dst.GoStmt); ok {
			return false
		}
		return !prevAssignsUsedVar(prev, curr)
	}

	// All remaining statements (if/for/range/switch/typeSwitch/select/expr/send):
	// only cuddle if prev assigns/declares a variable used in curr.
	return !prevAssignsUsedVar(prev, curr)
}

// prevAssignsUsedVar returns true if prev assigns or declares at least one variable
// that is used in curr.
func prevAssignsUsedVar(prev, curr dst.Stmt) bool {
	prevVars := extractAllVars(prev)
	if len(prevVars) == 0 {
		return false
	}
	return usesAnyIdent(curr, prevVars)
}

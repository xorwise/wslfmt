package fixer

import (
	"bytes"
	"go/token"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
)

// Fix parses src as a Go source file, applies WSL fixes, and returns the result.
func Fix(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, "", src, 0)
	if err != nil {
		return nil, err
	}

	dstutil.Apply(f, func(cursor *dstutil.Cursor) bool {
		block, ok := cursor.Node().(*dst.BlockStmt)
		if !ok {
			return true
		}
		fixBlock(block)
		return true
	}, nil)

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fixBlock applies all WSL block-level rules to the given block statement.
func fixBlock(block *dst.BlockStmt) {
	stmts := block.List
	if len(stmts) == 0 {
		return
	}

	// Rule: block should not start with a whitespace.
	setSpaceBefore(stmts[0], dst.NewLine)

	// Rule: block should not end with a whitespace.
	removeTrailingSpace(stmts[len(stmts)-1])

	// Apply pair-wise rules.
	for i := 1; i < len(stmts); i++ {
		prev := stmts[i-1]
		curr := stmts[i]

		if shouldHaveBlankLine(prev, curr, len(stmts)) {
			setSpaceBefore(curr, dst.EmptyLine)
		} else {
			// Only remove blank line if it's not required for readability;
			// we don't force-cuddle — we only add blank lines, not remove them.
			// Exception: error check must be cuddled (no blank line) when conditions met.
			if isErrorCheckIf(curr) && isAssignment(prev) {
				assignedVars := extractAssignedIdents(prev)
				if usesAnyIdent(curr, assignedVars) {
					setSpaceBefore(curr, dst.NewLine)
				}
			}
		}
	}
}

// setSpaceBefore sets the Before decoration of a statement.
func setSpaceBefore(stmt dst.Stmt, space dst.SpaceType) {
	decs := stmtDecs(stmt)
	if decs != nil {
		decs.Before = space
	}
}

// removeTrailingSpace ensures no blank line at the end of a statement's scope.
func removeTrailingSpace(stmt dst.Stmt) {
	decs := stmtDecs(stmt)
	if decs != nil && decs.After == dst.EmptyLine {
		decs.After = dst.NewLine
	}
}

// stmtDecs returns a pointer to the NodeDecs of a statement.
func stmtDecs(stmt dst.Stmt) *dst.NodeDecs {
	switch s := stmt.(type) {
	case *dst.AssignStmt:
		return &s.Decs.NodeDecs
	case *dst.ExprStmt:
		return &s.Decs.NodeDecs
	case *dst.ReturnStmt:
		return &s.Decs.NodeDecs
	case *dst.IfStmt:
		return &s.Decs.NodeDecs
	case *dst.ForStmt:
		return &s.Decs.NodeDecs
	case *dst.RangeStmt:
		return &s.Decs.NodeDecs
	case *dst.SwitchStmt:
		return &s.Decs.NodeDecs
	case *dst.TypeSwitchStmt:
		return &s.Decs.NodeDecs
	case *dst.SelectStmt:
		return &s.Decs.NodeDecs
	case *dst.BranchStmt:
		return &s.Decs.NodeDecs
	case *dst.DeclStmt:
		return &s.Decs.NodeDecs
	case *dst.DeferStmt:
		return &s.Decs.NodeDecs
	case *dst.GoStmt:
		return &s.Decs.NodeDecs
	case *dst.SendStmt:
		return &s.Decs.NodeDecs
	case *dst.IncDecStmt:
		return &s.Decs.NodeDecs
	case *dst.LabeledStmt:
		return &s.Decs.NodeDecs
	case *dst.BlockStmt:
		return &s.Decs.NodeDecs
	case *dst.EmptyStmt:
		return &s.Decs.NodeDecs
	}
	return nil
}

package fixer

import (
	"strings"
	"testing"
)

func testFix(t *testing.T, name, input, want string) {
	t.Helper()
	got, err := Fix([]byte(input))
	if err != nil {
		t.Fatalf("%s: Fix error: %v", name, err)
	}
	gotStr := strings.TrimSpace(string(got))
	wantStr := strings.TrimSpace(want)
	if gotStr != wantStr {
		t.Errorf("%s:\ngot:\n%s\n\nwant:\n%s", name, gotStr, wantStr)
	}
}

func TestBlockStartWhitespace(t *testing.T) {
	input := `package main

func f() {

	x := 1
	_ = x
}
`
	want := `package main

func f() {
	x := 1
	_ = x
}`
	testFix(t, "block-start", input, want)
}

func TestBlockEndWhitespace(t *testing.T) {
	input := `package main

func f() {
	x := 1
	_ = x

}
`
	want := `package main

func f() {
	x := 1
	_ = x
}`
	testFix(t, "block-end", input, want)
}

func TestIfCuddleWithAssignment(t *testing.T) {
	input := `package main

func f() {
	x := getValue()
	if x > 0 {
		_ = x
	}
}
`
	want := `package main

func f() {
	x := getValue()
	if x > 0 {
		_ = x
	}
}`
	testFix(t, "if-cuddle-with-assign", input, want)
}

func TestIfNoCuddleWithNonAssignment(t *testing.T) {
	input := `package main

func f() {
	doSomething()
	if true {
		doSomething()
	}
}
`
	want := `package main

func f() {
	doSomething()

	if true {
		doSomething()
	}
}`
	testFix(t, "if-no-cuddle-with-non-assign", input, want)
}

func TestErrorCheckCuddle(t *testing.T) {
	input := `package main

func f() error {
	x, err := doSomething()
	if err != nil {
		return err
	}
	_ = x
	return nil
}
`
	want := `package main

func f() error {
	x, err := doSomething()
	if err != nil {
		return err
	}

	_ = x

	return nil
}`
	testFix(t, "error-check-cuddle", input, want)
}

func TestReturnNoCuddleInLargeBlock(t *testing.T) {
	input := `package main

func f() int {
	a := 1
	b := 2
	c := a + b
	return c
}
`
	want := `package main

func f() int {
	a := 1
	b := 2
	c := a + b

	return c
}`
	testFix(t, "return-no-cuddle-large-block", input, want)
}

func TestReturnCuddleInSmallBlock(t *testing.T) {
	input := `package main

func f() int {
	a := 1
	return a
}
`
	want := `package main

func f() int {
	a := 1
	return a
}`
	testFix(t, "return-cuddle-small-block", input, want)
}

func TestBlankLineAfterBlock(t *testing.T) {
	input := `package main

func f() {
	if true {
		doSomething()
	}
	doSomethingElse()
}
`
	want := `package main

func f() {
	if true {
		doSomething()
	}

	doSomethingElse()
}`
	testFix(t, "blank-after-block", input, want)
}

func TestAssignmentNotCuddledWithNonAssign(t *testing.T) {
	input := `package main

func f() {
	doSomething()
	x := 1
	_ = x
}
`
	want := `package main

func f() {
	doSomething()

	x := 1
	_ = x
}`
	testFix(t, "assign-not-cuddled-with-non-assign", input, want)
}

func TestAppendCuddleWithAppendedValue(t *testing.T) {
	// append cuddled with statement that uses the appended variable — OK, no blank line.
	input := `package main

func f() {
	var s []int
	s = append(s, 1)
	_ = s
}
`
	want := `package main

func f() {
	var s []int
	s = append(s, 1)
	_ = s
}`
	testFix(t, "append-cuddle-with-appended", input, want)
}

func TestAppendNoCuddleWithUnrelated(t *testing.T) {
	// append not cuddled with unrelated statement — blank line required.
	input := `package main

func f() {
	x := 1
	s = append(s, x)
	_ = s
}
`
	want := `package main

func f() {
	x := 1

	s = append(s, x)
	_ = s
}`
	testFix(t, "append-no-cuddle-unrelated", input, want)
}

func TestDeclNeverCuddled(t *testing.T) {
	input := `package main

func f() {
	x := 1
	var y int
	_ = x + y
}
`
	want := `package main

func f() {
	x := 1

	var y int

	_ = x + y
}`
	testFix(t, "decl-never-cuddled", input, want)
}

func TestLabelNeverCuddled(t *testing.T) {
	input := `package main

func f() {
	x := 1
	_ = x
loop:
	_ = x
}
`
	want := `package main

func f() {
	x := 1
	_ = x

loop:
	_ = x
}`
	testFix(t, "label-never-cuddled", input, want)
}

func TestIncDecCuddledWithAssign(t *testing.T) {
	input := `package main

func f() {
	i := 0
	i++
	_ = i
}
`
	want := `package main

func f() {
	i := 0
	i++
	_ = i
}`
	testFix(t, "inc-dec-cuddled-with-assign", input, want)
}

func TestIncDecNoCuddleWithExpr(t *testing.T) {
	input := `package main

func f() {
	doSomething()
	i++
}
`
	want := `package main

func f() {
	doSomething()

	i++
}`
	testFix(t, "inc-dec-no-cuddle-with-expr", input, want)
}

func TestSwitchCuddleWithUsedVar(t *testing.T) {
	input := `package main

func f() {
	x := getValue()
	switch x {
	case 1:
		doSomething()
	}
}
`
	want := `package main

func f() {
	x := getValue()
	switch x {
	case 1:
		doSomething()
	}
}`
	testFix(t, "switch-cuddle-with-used-var", input, want)
}

func TestSwitchNoCuddleWithUnrelated(t *testing.T) {
	input := `package main

func f() {
	doSomething()
	switch x {
	case 1:
		doSomething()
	}
}
`
	want := `package main

func f() {
	doSomething()

	switch x {
	case 1:
		doSomething()
	}
}`
	testFix(t, "switch-no-cuddle-with-unrelated", input, want)
}

func TestDeferCuddledWithOtherDefer(t *testing.T) {
	input := `package main

func f() {
	defer cleanup1()
	defer cleanup2()
}
`
	want := `package main

func f() {
	defer cleanup1()
	defer cleanup2()
}`
	testFix(t, "defer-cuddled-with-defer", input, want)
}

func TestDeferCuddledWithUsedVar(t *testing.T) {
	input := `package main

func f() {
	f, err := os.Open("file")
	if err != nil {
		return
	}
	defer f.Close()
}
`
	want := `package main

func f() {
	f, err := os.Open("file")
	if err != nil {
		return
	}

	defer f.Close()
}`
	testFix(t, "defer-cuddle-with-used-var", input, want)
}

func TestGoCuddledWithOtherGo(t *testing.T) {
	input := `package main

func f() {
	go worker1()
	go worker2()
}
`
	want := `package main

func f() {
	go worker1()
	go worker2()
}`
	testFix(t, "go-cuddled-with-go", input, want)
}

func TestSendCuddledWithUsedVar(t *testing.T) {
	input := `package main

func f(ch chan int) {
	x := getValue()
	ch <- x
}
`
	want := `package main

func f(ch chan int) {
	x := getValue()
	ch <- x
}`
	testFix(t, "send-cuddle-with-used-var", input, want)
}

func TestIfNoCuddleWhenVarUsedOnlyInBody(t *testing.T) {
	// x is assigned by prev but used only in the if body, not the condition.
	// WSL: "if statements should only be cuddled with assignments used in the if statement itself"
	input := `package main

func f() {
	x := 1
	if y > 0 {
		_ = x
	}
}
`
	want := `package main

func f() {
	x := 1

	if y > 0 {
		_ = x
	}
}`
	testFix(t, "if-no-cuddle-var-only-in-body", input, want)
}

func TestSendNoCuddleWithUnrelated(t *testing.T) {
	input := `package main

func f(ch chan int) {
	doSomething()
	ch <- 1
}
`
	want := `package main

func f(ch chan int) {
	doSomething()

	ch <- 1
}`
	testFix(t, "send-no-cuddle-unrelated", input, want)
}

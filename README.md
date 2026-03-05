# wslfmt

A formatter for Go that automatically fixes [WSL (Whitespace Linter)](https://github.com/bombsimon/wsl) errors using [`github.com/dave/dst`](https://github.com/dave/dst).

## Install

```sh
curl -sSfL https://raw.githubusercontent.com/xorwise/wslfmt/main/install.sh | sh
```

Or with `go install`:

```sh
go install github.com/xorwise/wslfmt/cmd/wslfmt@latest
```

## Usage

```sh
# Fix files in place
wslfmt -w ./...

# Print fixed output to stdout
wslfmt file.go

# List files that differ
wslfmt -l ./...

# Show diff
wslfmt -d ./...
```

## Editor Integration

### VS Code / Cursor

Install the [Run on Save](https://marketplace.visualstudio.com/items?itemName=pucelle.run-on-save) extension, then add to `settings.json`:

```json
"runOnSave.onlyRunOnManualSave": true,
"runOnSave.commands": [
    {
        "match": ".*\\.go$",
        "command": "wslfmt -w ${file}",
        "runIn": "backend"
    }
]
```

`wslfmt` will run automatically every time you manually save a `.go` file.

## Rules

| Rule | Description |
|------|-------------|
| `assign` | Assignments only cuddle with other assignments or inc/dec |
| `branch` | break/continue/return require blank line if block > 2 lines |
| `decl` | Declarations (`var`/`const`/`type`) are never cuddled |
| `defer` | Defer cuddles with other defer or a variable used on the line above |
| `expr` | Expressions only cuddle with variables used on the line above |
| `for` / `range` | Only cuddle with a variable used on the line above |
| `go` | Cuddles with other go or a variable used on the line above |
| `if` | Only cuddles with a variable used on the line above |
| `inc-dec` | Same rules as assign |
| `label` | Labels are never cuddled |
| `select` / `switch` / `type-switch` | Only cuddle with a variable used on the line above |
| `send` | Only cuddles with a variable used on the line above |
| `append` | `x = append(x, v)` only cuddles with the line that uses/declares `x` |
| `err` | `if err != nil` is force-cuddled with the assignment that set `err` |
| `leading-whitespace` | No blank line at the start of a block |
| `trailing-whitespace` | No blank line at the end of a block |

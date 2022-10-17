[![GoDoc](https://godoc.org/github.com/baldurstod/vdf?status.png)](http://godoc.org/github.com/baldurstod/vdf)
[![Go Report Card](https://goreportcard.com/badge/github.com/baldurstod/vdf)](https://goreportcard.com/badge/github.com/baldurstod/vdf)

# vdf
A VDF (Valve Data Format) parser for go. Handles comments and UTF-8 characters

# Installation

```
go get -u github.com/baldurstod/vdf
```

# Example

Parse data inside `items_game.txt`

```go
package main

import (
	"os"
	"github.com/baldurstod/vdf"
)

func main() {
	buf, err := os.ReadFile("items_game.txt")
	if err != nil {
		panic(err)
	}

	vdf := vdf.VDF{}
	items := vdf.Parse(buf)

	/*
	 ...
	 */
}
```

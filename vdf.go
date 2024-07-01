package vdf

import (
	"fmt"
	"unicode/utf8"

	"github.com/golang-collections/collections/stack"
)

type Token int

const (
	OPENING_BRACE Token = iota
	CLOSING_BRACE
	NEW_LINE
	STRING_VALUE
	END_TOKEN
)

type VDF struct {
	s   []byte
	i   int
	len int
}

func PrintTabs(tabs int) {
	for i := 0; i < tabs; i++ {
		fmt.Print("\t")
	}
}

func (vdf *VDF) Parse(s []byte) KeyValue {
	vdf.s = s
	vdf.i = 0
	vdf.len = len(s)

	stringStack := stack.New()
	levelStack := stack.New()

	var currentLevel *KeyValue = &KeyValue{Key: "root", Value: []*KeyValue{}, isRoot: true}
	var result KeyValue

TokenLoop:
	for {
		token, s := vdf.getNextToken()
		switch token {
		case OPENING_BRACE:
			key := stringStack.Pop().(string)
			subLevel := KeyValue{Key: key, Value: []*KeyValue{}}

			if currentLevel != nil {
				currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &subLevel)
			}

			levelStack.Push(currentLevel)
			currentLevel = &subLevel
		case CLOSING_BRACE:
			currentLevel = levelStack.Pop().(*KeyValue)
			if currentLevel != nil {
				result = *currentLevel
			}
		case NEW_LINE:
			if stringStack.Len() > 1 {
				value := stringStack.Pop().(string)
				key := stringStack.Pop().(string)
				currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &KeyValue{Key: key, Value: value})
			}
		case STRING_VALUE:
			stringStack.Push(s)
		case END_TOKEN:
			break TokenLoop
		}
	}

	return result
}

func (vdf *VDF) getNextRune() (rune, int) {
	c, size := utf8.DecodeRune(vdf.s)
	vdf.s = vdf.s[size:]

	return c, size
}

func (vdf *VDF) getNextToken() (Token, string) {
	for vdf.i < vdf.len {
		c, size := vdf.getNextRune()
		vdf.i += size
		switch c {
		case '{':
			return OPENING_BRACE, ""
		case '}':
			return CLOSING_BRACE, ""
		case '\r', '\n':
			return NEW_LINE, ""
		case ' ', '\t': //just eat a char
		case '"':
			s := ""
			for vdf.i < vdf.len {
				c, size := vdf.getNextRune()
				vdf.i += size
				switch c {
				case '\\':
					if vdf.i < vdf.len {
						c, size := vdf.getNextRune()
						vdf.i += size
						if c == '"' {
							s += "\\\""
						} else {
							s += `\` + string(c)
						}
					}
				case '"':
					return STRING_VALUE, s
				default:
					s += string(c)
				}
			}
		case '/':
			for vdf.i < vdf.len {
				c, size := vdf.getNextRune()
				vdf.i += size
				if c == '\r' || c == '\n' {
					break
				}
			}
		}
	}
	return END_TOKEN, ""
}

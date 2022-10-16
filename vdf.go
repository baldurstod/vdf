package vdf

import (
	"github.com/golang-collections/collections/stack"
	"unicode/utf8"
)

var depth = 0

type Token int

const (
	openingBrace Token = iota
	closingBrace
	newLine
	stringValue
	endToken
)

type VDF struct {
	s []byte
	i int
	len int
}


func (this *VDF) Parse(s []byte) map[string]interface{} {
	this.s = s
	this.i = 0
	this.len = len(s)
	result := map[string]interface{}{}

	stringStack := stack.New()
	levelStack := stack.New()
	currentLevel := result

TokenLoop:
	for {
		token, s := this.getNextToken()
		switch token {
		case openingBrace:
			key := stringStack.Pop().(string)
			var subLevel map[string]interface{}
			if sl, exist := currentLevel[key]; exist {
				subLevel = sl.(map[string]interface{})
			} else {
				subLevel = map[string]interface{}{}
			}
			currentLevel[key] = subLevel
			levelStack.Push(currentLevel)
			currentLevel = subLevel
		case closingBrace:
			currentLevel = levelStack.Pop().(map[string]interface{})
		case newLine:
			if stringStack.Len() > 1 {
				value := stringStack.Pop().(string)
				key := stringStack.Pop().(string)
				currentLevel[key] = value
			}
		case stringValue: stringStack.Push(s)
		case endToken: break TokenLoop
		}

	}

	return result
}

func (this *VDF) getNextRune() (rune, int) {
	c, size := utf8.DecodeRune(this.s)
	this.s = this.s[size:]

	return c, size
}

func (this *VDF) getNextToken() (Token, string) {
	for this.i < this.len {
		c, size := this.getNextRune()
		this.i += size
		switch c {
		case '{': return openingBrace, ""
		case '}': return closingBrace, ""
		case '\r', '\n': return newLine, ""
		case ' ', '\t'://just eat a char
		case '"':
			s := ""
			for this.i < this.len {
				c, size := this.getNextRune()
				this.i += size
				switch c {
				case '\\':
					if this.i < this.len {
						c, size := this.getNextRune()
						this.i += size
						if c == '"' {
							s += "\\\""
						} else {
							s += `\` + string(c)
						}
					}
				case '"':
					return stringValue, s
				default:
					s += string(c)
				}
			}
		case '/':
			for this.i < this.len {
				c, size := this.getNextRune()
				this.i += size
				if c == '\r' || c == '\n' {
					break
				}
			}
		}
	}
	return endToken, ""
}

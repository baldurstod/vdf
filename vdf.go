package vdf

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/golang-collections/collections/stack"
)

type Token int

const (
	INVALID_TOKEN Token = iota
	OPENING_BRACE
	CLOSING_BRACE
	NEW_LINE
	STRING_VALUE
	BASE
	END_TOKEN
)

type VDF struct {
	s   []byte
	i   int
	len int
	t   Token
}

func PrintTabs(tabs int) {
	for i := 0; i < tabs; i++ {
		fmt.Print("\t")
	}
}

type GetFileContent func(filepath string) ([]byte, error)

func (vdf *VDF) Parse(s []byte, getFileContent GetFileContent) KeyValue {
	vdf.s = s
	vdf.i = 0
	vdf.len = len(s)
	vdf.t = INVALID_TOKEN

	stringStack := stack.New()
	levelStack := stack.New()

	var currentLevel *KeyValue = &KeyValue{Key: "root" /*, Value: []*KeyValue{}*/, isRoot: true}
	var result KeyValue
	var inBase = false

TokenLoop:
	for {
		token, s := vdf.getNextToken()
		switch token {
		case OPENING_BRACE:
			key := stringStack.Pop().(string)
			subLevel := KeyValue{Key: key /*, Value: []*KeyValue{}*/}

			if currentLevel != nil {
				//currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &subLevel)
				currentLevel.AddSubElement(&subLevel)
			}

			levelStack.Push(currentLevel)
			currentLevel = &subLevel
		case CLOSING_BRACE:
			currentLevel = levelStack.Pop().(*KeyValue)
			if currentLevel != nil {
				result = *currentLevel
			}
		case NEW_LINE:
			inBase = false
			if stringStack.Len() > 1 {
				value := stringStack.Pop().(string)
				key := stringStack.Pop().(string)
				//currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &KeyValue{Key: key, Value: value})
				stringValue := KeyValue{Key: key}
				stringValue.SetStringValue((value))
				currentLevel.AddSubElement(&stringValue)
			}
		case STRING_VALUE:
			if inBase {
				inBase = false
				if getFileContent != nil {
					data, err := getFileContent(s)
					if err == nil {
						vdf := VDF{}
						root := vdf.Parse(data, getFileContent)
						currentLevel.Merge((&root))
					}
				}
			} else {
				stringStack.Push(s)
			}
		case BASE:
			inBase = true
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

// Compare the next runes with the provided string. Only advance the cursor if th comparaison is true
func (vdf *VDF) compareNextString(s string) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	var b []byte = vdf.s
	totalSize := 0

	for i := 0; i < l; i++ {
		rune, size := utf8.DecodeRune(b)
		if string(rune) != s[i:i+1] {
			return false
		}
		b = b[size:]
		totalSize += size
	}

	vdf.s = b
	vdf.i += totalSize

	return true
}

/*
// Pick the next `len` runes. Doesn't modify the content
func (vdf *VDF) pickNextRunes(len int) []rune {
	if len <= 0 {
		return nil
	}

	var b []byte = vdf.s
	var size int
	runes := make([]rune, len)

	for i := 0; i < len; i++ {
		runes[i], size = utf8.DecodeRune(b)
		b = b[size:]
	}

	return runes
}
*/

func (vdf *VDF) getNextToken() (Token, string) {
	if vdf.t != INVALID_TOKEN {
		t := vdf.t
		vdf.t = INVALID_TOKEN
		return t, ""
	}

	var sb strings.Builder

	for vdf.i < vdf.len {
		c, size := vdf.getNextRune()
		// Safe guard if we go past the end
		if size == 0 {
			return END_TOKEN, ""
		}
		vdf.i += size
		switch c {
		case '{':
			return OPENING_BRACE, ""
		case '}':
			return CLOSING_BRACE, ""
		case '\r', '\n':
			if sb.Len() != 0 {
				vdf.t = NEW_LINE
				return STRING_VALUE, sb.String()
			} else {
				return NEW_LINE, ""
			}
		case ' ', '\t': //just eat a char
		case '"':
			var sb strings.Builder
			for vdf.i < vdf.len {
				c, size := vdf.getNextRune()
				vdf.i += size
				switch c {
				case '\\':
					if vdf.i < vdf.len {
						c, size := vdf.getNextRune()
						vdf.i += size
						if c == '"' {
							sb.WriteString("\\\"")
						} else {
							sb.WriteString(`\`)
							sb.WriteString(string(c))
						}
					}
				case '"':
					return STRING_VALUE, sb.String()
				default:
					sb.WriteString(string(c))
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
		case '#':
			// If we have a #, check if we are at the start of  a #base instruction
			if vdf.compareNextString("base") {
				return BASE, ""
			} else {
				sb.WriteString(string(c))
			}
		default:
			sb.WriteString(string(c))
		}
	}
	return END_TOKEN, ""
}

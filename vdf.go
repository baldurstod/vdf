package vdf

import (
	"encoding/json"
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

type KeyValue struct {
	Key    string
	Value  interface{}
	isRoot bool
}

func PrintTabs(tabs int) {
	for i := 0; i < tabs; i++ {
		fmt.Print("\t")
	}
}

func (kv *KeyValue) GetString(key string) (string, bool) {
	a, ok := kv.Get(key)
	if ok {
		switch a.Value.(type) {
		case string:
			return a.Value.(string), true
		}
	}
	return "", false
}

func (kv *KeyValue) ToString() (string, bool) {
	switch kv.Value.(type) {
	case string:
		return kv.Value.(string), true
	}
	return "", false
}

func (kv *KeyValue) Get(key string) (*KeyValue, bool) {
	switch kv.Value.(type) {
	case string:
		return nil, false
	case []*KeyValue:
		arr := kv.Value.([]*KeyValue)
		for _, item := range arr {
			if key == item.Key {
				return item, true
			}
		}
	}
	return nil, false
}

func (kv *KeyValue) GetAll(key string) ([]*KeyValue, bool) {
	switch kv.Value.(type) {
	case string:
		return nil, false
	case []*KeyValue:
		ret := []*KeyValue{}
		arr := kv.Value.([]*KeyValue)
		for _, item := range arr {
			if key == item.Key {
				ret = append(ret, item)
			}
		}
		if len(ret) > 0 {
			return ret, true
		}
	}
	return nil, false
}

func (kv *KeyValue) GetSubElement(path []string) (*KeyValue, bool) {
	if subElements, ok := kv.GetAll(path[0]); ok {
		if len(path) == 1 {
			return subElements[0], true
		} else {
			for _, subElement := range subElements {
				if subElement2, ok := subElement.GetSubElement(path[1:]); ok {
					return subElement2, true
				}
			}
		}
	}
	return nil, false
}

func (kv *KeyValue) GetChilds() []*KeyValue {
	switch kv.Value.(type) {
	case []*KeyValue:
		return kv.Value.([]*KeyValue)
	}
	return []*KeyValue{}
}

func (kv *KeyValue) ToStringMap() (*map[string]string, bool) {
	switch kv.Value.(type) {
	case []*KeyValue:
		ret := make(map[string]string)
		arr := kv.Value.([]*KeyValue)
		for _, item := range arr {
			switch item.Value.(type) {
			case string:
				ret[item.Key] = item.Value.(string)
			}
		}
		return &ret, true
	}
	return nil, false
}

func (kv *KeyValue) GetStringMap(key string) (*map[string]string, bool) {
	if sub, ok := kv.Get(key); ok {
		return sub.ToStringMap()
	}
	return nil, false
}

func (kv *KeyValue) GetSubElementStringMap(path []string) (*map[string]string, bool) {
	if sub, ok := kv.GetSubElement(path); ok {
		return sub.ToStringMap()
	}
	return nil, false
}

func (kv *KeyValue) RemoveDuplicates() {
	switch kv.Value.(type) {
	case []*KeyValue:
		allKeys := make(map[string]bool)
		list := []*KeyValue{}

		arr := kv.Value.([]*KeyValue)
		for _, item := range arr {
			key := item.Key
			if _, value := allKeys[key]; !value {
				allKeys[key] = true
				list = append(list, item)
				item.RemoveDuplicates()
			}
		}
		kv.Value = list
	}
}

func (kv *KeyValue) Print(optional ...int) {
	tabs := 0
	if len(optional) > 0 {
		tabs = optional[0]
	}

	if kv.isRoot {
		tabs = -1
	}

	switch kv.Value.(type) {
	case []*KeyValue:
		if !kv.isRoot {
			PrintTabs(tabs)
			fmt.Println("\"" + kv.Key + "\"")
			PrintTabs(tabs)
			fmt.Println("{")
		}
		arr := kv.Value.([]*KeyValue)
		for _, val := range arr {
			val.Print(tabs + 1)
		}

		if !kv.isRoot {
			PrintTabs(tabs)
			fmt.Println("}")
		}
	case string:
		PrintTabs(tabs)
		fmt.Println("\"" + kv.Key + "\"		\"" + kv.Value.(string) + "\"")
	default:
		fmt.Println(kv)
		panic("unknown type")
	}
}

func (kv *KeyValue) toJSON() interface{} {
	ret := make(map[string]interface{})

	switch kv.Value.(type) {
	case string:
		return kv.Value.(string)
	case []*KeyValue:
		arr := kv.Value.([]*KeyValue)
		for _, subKv := range arr {
			ret[subKv.Key] = subKv.toJSON()
		}
	}

	return ret
}

func (kv *KeyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(kv.toJSON())
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

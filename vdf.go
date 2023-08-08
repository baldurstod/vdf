package vdf

import (
	"github.com/golang-collections/collections/stack"
	"unicode/utf8"
	"fmt"
	"encoding/json"
)

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

type KeyValue struct {
	Key string
	Value interface{}
	isRoot bool `default:false`
}

func PrintTabs(tabs int) {
	for i := 0; i < tabs; i++ {
		fmt.Print("\t")
	}
}

func (this KeyValue) GetString(key string) (string, bool) {
	a, ok := this.Get(key)
	if ok {
		switch a.Value.(type) {
		case string:
			return a.Value.(string), true
		}
	}
	return "", false
}

func (this KeyValue) ToString() (string, bool) {
	switch this.Value.(type) {
	case string:
		return this.Value.(string), true
	}
	return "", false
}

func (this KeyValue) Get(key string) (*KeyValue, bool) {
	switch this.Value.(type) {
	case string:
		return nil, false
	case []*KeyValue:
		arr := this.Value.([]*KeyValue)
		for _, item := range arr {
			if key == item.Key {
				return item, true
			}
		}
	}
	return nil, false
}

func (this KeyValue) GetAll(key string) ([]*KeyValue, bool) {
	switch this.Value.(type) {
	case string:
		return nil, false
	case []*KeyValue:
		ret := []*KeyValue{}
		arr := this.Value.([]*KeyValue)
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

func (this KeyValue) GetSubElement(path []string) (*KeyValue, bool) {
	if subElements, ok := this.GetAll(path[0]); ok {
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

func (this KeyValue) GetChilds() ([]*KeyValue) {
	switch this.Value.(type) {
	case []*KeyValue:
		return this.Value.([]*KeyValue)
	}
	return []*KeyValue{}
}

func (this KeyValue) ToStringMap() (*map[string]string, bool) {
	switch this.Value.(type) {
	case []*KeyValue:
		ret := make(map[string]string)
		arr := this.Value.([]*KeyValue)
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

func (this KeyValue) GetStringMap(key string) (*map[string]string, bool) {
	if sub, ok := this.Get(key); ok {
		return sub.ToStringMap()
	}
	return nil, false
}

func (this KeyValue) GetSubElementStringMap(path []string) (*map[string]string, bool) {
	if sub, ok := this.GetSubElement(path); ok {
		return sub.ToStringMap()
	}
	return nil, false
}

func (this KeyValue) RemoveDuplicates() {
	switch this.Value.(type) {
	case []*KeyValue:
		allKeys := make(map[string]bool)
		list := []*KeyValue{}

		arr := this.Value.([]*KeyValue)
		for _, item := range arr {
			key := item.Key
			if _, value := allKeys[key]; !value {
				allKeys[key] = true
				list = append(list, item)
				item.RemoveDuplicates()
			}
		}
		this.Value = list
	}
}

func (this KeyValue) Print(optional ...int) {
	tabs := 0
	if len(optional) > 0 {
		tabs = optional[0]
	}

	if this.isRoot {
		tabs = -1
	}

	switch this.Value.(type) {
	case []*KeyValue:
		if !this.isRoot {
			PrintTabs(tabs)
			fmt.Println("\"" + this.Key + "\"")
			PrintTabs(tabs)
			fmt.Println("{")
		}
		arr := this.Value.([]*KeyValue)
		for _, val := range arr {
			val.Print(tabs + 1);
		}

		if !this.isRoot {
			PrintTabs(tabs)
			fmt.Println("}")
		}
	case string:
		PrintTabs(tabs)
		fmt.Println("\"" + this.Key + "\"		\"" + this.Value.(string) + "\"")
	default:
		fmt.Println(this)
		panic("unknown type")
	}
}

func (this *KeyValue) toJSON() interface{} {
	ret := make(map[string]interface{})

	switch this.Value.(type) {
	case string:
		return this.Value.(string)
	case []*KeyValue:
		arr := this.Value.([]*KeyValue)
		for _, kv := range arr {
			ret[kv.Key] = kv.toJSON()
		}
	}

	return ret
}

func (this KeyValue) MarshalJSON() ([]byte, error) {
	var ret interface{}

	ret = this.toJSON()

	return json.Marshal(ret)
}


func (this *VDF) Parse(s []byte) KeyValue {
	this.s = s
	this.i = 0
	this.len = len(s)

	stringStack := stack.New()
	levelStack := stack.New()

	var currentLevel *KeyValue = &KeyValue{Key: "root", Value: []*KeyValue{}, isRoot: true}
	var result KeyValue

TokenLoop:
	for {
		token, s := this.getNextToken()
		switch token {
		case openingBrace:
			key := stringStack.Pop().(string)
			subLevel := KeyValue{Key: key, Value: []*KeyValue{}}

			if currentLevel != nil {
				currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &subLevel)
			}

			levelStack.Push(currentLevel)
			currentLevel = &subLevel
		case closingBrace:
			currentLevel = levelStack.Pop().(*KeyValue)
			if currentLevel != nil {
				result = *currentLevel
			}
		case newLine:
			if stringStack.Len() > 1 {
				value := stringStack.Pop().(string)
				key := stringStack.Pop().(string)
				currentLevel.Value = append(currentLevel.Value.([]*KeyValue), &KeyValue{Key: key, Value: value})
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

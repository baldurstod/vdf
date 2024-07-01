package vdf

import (
	"encoding/json"
	"fmt"
)

type KeyValue struct {
	Key    string
	Value  interface{}
	isRoot bool
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

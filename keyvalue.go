package vdf

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type KeyValue struct {
	Key    string
	Value  interface{}
	isRoot bool
}

func (kv *KeyValue) GetString(key string) (string, error) {
	a, err := kv.Get(key)
	if err != nil {
		return "", err
	}

	s, ok := a.Value.(string)
	if !ok {
		return "", errors.New("unexpected value type for key " + key)
	}

	return s, nil
}

func (kv *KeyValue) GetInt(key string) (int, error) {
	a, err := kv.Get(key)
	if err != nil {
		return 0, err
	}

	s, ok := a.Value.(string)
	if !ok {
		return 0, errors.New("unexpected value type for key " + key)
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("can't convert key %s to int: <%w>", key, err)
	}

	return i, nil
}

func (kv *KeyValue) GetBool(key string) (bool, error) {
	s, err := kv.GetString(key)
	if err != nil {
		return false, err
	}

	if s == "1" {
		return true, nil
	} else {
		return false, nil
	}
}

func (kv *KeyValue) ToString() (string, error) {
	switch kv.Value.(type) {
	case string:
		return kv.Value.(string), nil
	default:
		return "", errors.New("unexpected value type")
	}
}

func (kv *KeyValue) Get(key string) (*KeyValue, error) {
	switch v := kv.Value.(type) {
	case []*KeyValue:
		for _, item := range v {
			if key == item.Key {
				return item, nil
			}
		}
	default:
		return nil, errors.New("unexpected element type")
	}
	return nil, errors.New("key not found: " + key)
}

func (kv *KeyValue) GetAll(key string) ([]*KeyValue, error) {
	switch v := kv.Value.(type) {
	case []*KeyValue:
		ret := []*KeyValue{}
		for _, item := range v {
			if key == item.Key {
				ret = append(ret, item)
			}
		}
		return ret, nil
	default:
		return nil, errors.New("unexpected element type")
	}
}

func (kv *KeyValue) GetSubElement(path []string) (*KeyValue, error) {
	subElements, err := kv.GetAll(path[0])
	if err != nil {
		return nil, err
	}

	if len(path) == 1 {
		return subElements[0], nil
	} else {
		for _, subElement := range subElements {
			if subElement2, err := subElement.GetSubElement(path[1:]); err == nil {
				return subElement2, nil
			}
		}
	}

	return nil, errors.New("subelement not found for path: " + strings.Join(path, "."))
}

func (kv *KeyValue) GetChilds() []*KeyValue {
	switch kv.Value.(type) {
	case []*KeyValue:
		return kv.Value.([]*KeyValue)
	}
	return []*KeyValue{}
}

func (kv *KeyValue) ToStringMap() (*map[string]string, error) {
	switch v := kv.Value.(type) {
	case []*KeyValue:
		ret := make(map[string]string)
		for _, item := range v {
			switch item.Value.(type) {
			case string:
				ret[item.Key] = item.Value.(string)
			}
		}
		return &ret, nil
	default:
		return nil, errors.New("unexpected element type")
	}
}

func (kv *KeyValue) GetStringMap(key string) (*map[string]string, error) {
	sub, err := kv.Get(key)

	if err != nil {
		return nil, err
	}
	return sub.ToStringMap()
}

func (kv *KeyValue) GetSubElementStringMap(path []string) (*map[string]string, error) {
	sub, err := kv.GetSubElement(path)
	if err != nil {
		return nil, err
	}
	return sub.ToStringMap()
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

	switch v := kv.Value.(type) {
	case []*KeyValue:
		if !kv.isRoot {
			PrintTabs(tabs)
			fmt.Println("\"" + kv.Key + "\"")
			PrintTabs(tabs)
			fmt.Println("{")
		}
		for _, val := range v {
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

	switch v := kv.Value.(type) {
	case string:
		return kv.Value.(string)
	case []*KeyValue:
		for _, subKv := range v {
			ret[subKv.Key] = subKv.toJSON()
		}
	}

	return ret
}

func (kv *KeyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(kv.toJSON())
}

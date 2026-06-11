package vdf

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type KeyValue struct {
	Key string
	// Value can be of type string or map[string][]*KeyValue
	value  any
	isRoot bool
}

func (kv *KeyValue) SetStringValue(value string) error {
	switch kv.value.(type) {
	case string:
	case nil:
	default:
		return errors.New("can't replace a non string value with a sting")
	}
	kv.value = value
	return nil
}

func (kv *KeyValue) AddSubElement(element *KeyValue) error {
	switch kv.value.(type) {
	case string:
		return errors.New("can't add a subelement to a string value")
	case nil:
		// Allocate the value
		kv.value = map[string][]*KeyValue{}
	}

	value, found := kv.value.(map[string][]*KeyValue)[element.Key]
	if found {
		kv.value.(map[string][]*KeyValue)[element.Key] = append(value, element)
	} else {
		kv.value.(map[string][]*KeyValue)[element.Key] = []*KeyValue{element}
	}

	return nil
}

func (kv *KeyValue) GetValue() any {
	return kv.value
}

func (kv *KeyValue) GetString(key string) (string, error) {
	a, err := kv.Get(key)
	if err != nil {
		return "", err
	}

	s, ok := a.value.(string)
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

	s, ok := a.value.(string)
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

func (kv *KeyValue) GetFloat32(key string) (float32, error) {
	s, err := kv.GetString(key)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, fmt.Errorf("can't convert key %s to float: <%w>", key, err)
	}

	return float32(f), nil
}

func (kv *KeyValue) GetFloat64(key string) (float64, error) {
	s, err := kv.GetString(key)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("can't convert key %s to float: <%w>", key, err)
	}

	return f, nil
}

func (kv *KeyValue) ToString() (string, error) {
	switch kv.value.(type) {
	case string:
		return kv.value.(string), nil
	default:
		return "", errors.New("unexpected value type")
	}
}

func (kv *KeyValue) ToInt() (int, error) {
	s, err := kv.ToString()
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("can't convert to int: <%w>", err)
	}

	return i, nil
}

func (kv *KeyValue) ToBool() (bool, error) {
	s, err := kv.ToString()
	if err != nil {
		return false, err
	}

	if s == "1" {
		return true, nil
	} else {
		return false, nil
	}
}

func (kv *KeyValue) Get(key string) (*KeyValue, error) {
	switch v := kv.value.(type) {
	case map[string][]*KeyValue:
		subElement, found := v[key]
		if found {
			return subElement[0], nil
		} else {
			return nil, errors.New("key not found: " + key)
		}
	default:
		return nil, errors.New("unexpected element type")
	}
}

func (kv *KeyValue) GetAll(key string) ([]*KeyValue, error) {
	switch v := kv.value.(type) {
	case map[string][]*KeyValue:
		subElement := v[key]
		return subElement, nil
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
	switch v := kv.value.(type) {
	case map[string][]*KeyValue:
		result := []*KeyValue{}

		for _, subKv := range v {
			result = append(result, subKv...)
		}

		return result
	}
	return []*KeyValue{}
}

func (kv *KeyValue) ToStringMap() (*map[string]string, error) {
	switch v := kv.value.(type) {
	case map[string][]*KeyValue:
		ret := make(map[string]string)
		for _, arr := range v {
			for _, item := range arr {
				switch item.value.(type) {
				case string:
					ret[item.Key] = item.value.(string)
				}
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

/*
func (kv *KeyValue) RemoveDuplicates() {
	switch kv.value.(type) {
	case []*KeyValue:
		allKeys := make(map[string]bool)
		list := []*KeyValue{}

		arr := kv.value.([]*KeyValue)
		for _, item := range arr {
			key := item.Key
			if _, value := allKeys[key]; !value {
				allKeys[key] = true
				list = append(list, item)
				item.RemoveDuplicates()
			}
		}
		kv.value = list
	}
}
*/

func (kv *KeyValue) Print(optional ...int) {
	tabs := 0
	if len(optional) > 0 {
		tabs = optional[0]
	}

	if kv.isRoot {
		tabs = -1
	}

	switch v := kv.value.(type) {
	case map[string][]*KeyValue:
		if !kv.isRoot {
			PrintTabs(tabs)
			fmt.Println("\"" + kv.Key + "\"")
			PrintTabs(tabs)
			fmt.Println("{")
		}
		for _, arr := range v {
			for _, val := range arr {
				val.Print(tabs + 1)
			}
		}

		if !kv.isRoot {
			PrintTabs(tabs)
			fmt.Println("}")
		}
	case string:
		PrintTabs(tabs)
		fmt.Println("\"" + kv.Key + "\"		\"" + kv.value.(string) + "\"")
	default:
		fmt.Println(kv)
		panic("unknown type")
	}
}

func (kv *KeyValue) toJSON() interface{} {
	ret := make(map[string]interface{})

	switch v := kv.value.(type) {
	case string:
		return kv.value.(string)
	case map[string][]*KeyValue:
		for key, subKv := range v {
			if len(subKv) == 1 {
				ret[key] = subKv[0].toJSON()
			} else {
				for i, element := range subKv {
					ret[key+"$"+strconv.Itoa(i)] = element.toJSON()
				}
			}
		}
	}

	return ret
}

func (kv *KeyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(kv.toJSON())
}

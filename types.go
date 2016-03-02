package transit_go

import "fmt"

type TaggedValue struct {
	Tag string
	Rep interface{}
}

type Quote struct {
	Object interface{}
}

type Keyword string
type Symbol string
type Tag string

/* ========== Link type ==================== */
type Link struct {
	Href   string
	Rel    string
	Name   string
	Prompt string
	Render string
}

func NewLink(href, rel, name, prompt, render string) (Link, error) {
	if href == "" {
		return Link{}, fmt.Errorf("Value of href cannot be empty")
	}
	if rel == "" {
		return Link{}, fmt.Errorf("Value of rel cannot be empty")
	}
	if render != "link" && render != "image" {
		return Link{}, fmt.Errorf("Value of render should be either 'link' or 'image'")
	}
	return Link{Href: href, Rel: rel, Name: name, Prompt: prompt, Render: render}, nil
}

func NewLinkFromMap(linkMap map[string]string) (Link, error) {
	return NewLink(linkMap["href"], linkMap["rel"], linkMap["name"], linkMap["prompt"], linkMap["render"])
}

/* ==========  Set type ==================== */

type Set interface {
	Add(elem interface{}) interface{}
	Contains(elem interface{}) bool
	Remove(elem interface{}) bool
	Items() []interface{}
	Len() int
}

type setStruct struct {
	backingMap map[*MapKey]bool
}

func NewSet() Set {
	return setStruct{backingMap: make(map[*MapKey]bool)}
}

func NewSetFrom(elements []interface{}) Set {
	set := NewSet()
	for _, elem := range elements {
		set.Add(elem)
	}
	return set
}

func (s setStruct) Add(elem interface{}) interface{} {
	if !s.Contains(elem) {
		s.backingMap[newMapKey(elem)] = true
	}
	return elem
}

func (s setStruct) Contains(elem interface{}) bool {
	for mapKey, _ := range s.backingMap {
		if mapKey.Key == elem {
			return true
		}
	}
	return false
}

func (s setStruct) Remove(elem interface{}) bool {
	for mapKey, _ := range s.backingMap {
		if mapKey.Key == elem {
			delete(s.backingMap, mapKey)
			return true
		}
	}
	return false
}

func (s setStruct) Items() []interface{} {
	var items []interface{}

	for mapKey, _ := range s.backingMap {
		items = append(items, mapKey.Key)
	}

	return items
}

func (s setStruct) Len() int {
	return len(s.backingMap)
}

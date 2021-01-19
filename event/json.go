package event

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.Config{
	EscapeHTML:                    false,
	SortMapKeys:                   false,
	UseNumber:                     false,
	ObjectFieldMustBeSimpleString: false,
}.Froze()

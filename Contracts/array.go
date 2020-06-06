package Contracts

type ArrayBase interface {
	Count() int
	Cap() int
	Empty() bool
}

type ArrayAccess interface {
	ArrayBase
	OffsetGet(interface{}) (interface{}, bool)
	OffsetSet(interface{}, interface{})
	OffsetExists(interface{}) bool
	OffsetUnset(interface{})
}

type ArrayIterator interface {
	Foreach(func(key, value interface{}) bool)
	Filter(func(key, value interface{}) bool) ArrayAccess
}

type ArrayAggregation interface {
	ArrayAccess
	ArrayIterator
	Stream(func(key, value interface{}) bool) ArrayAggregation
}

type UniqueArray interface {
	Exists(interface{}) bool
	Add(item interface{}) bool
}

package common

type Value interface {
	any
}

type ComparableValue interface {
	Value
	comparable
}

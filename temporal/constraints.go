package temporal

// Value represents any serializable value which may be converted to a payload.
type Value interface {
	any
}

// ComparableValue is a Value which may be compared to another value of the same
// type for equality.
type ComparableValue interface {
	Value
	comparable
}

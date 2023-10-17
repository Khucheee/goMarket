package app

type Operator struct {
	Field string
}

func NewOperator() *Operator {
	operator := Operator{
		Field: "JWT",
	}
	return &operator
}

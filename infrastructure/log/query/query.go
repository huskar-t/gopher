package query

type Ands map[string]interface{}

func And(tagName string, value interface{}) Ands {
	return Ands{
		tagName: value,
	}
}

func (ands Ands) And(tagName string, value interface{}) Ands {
	ands[tagName] = value
	return ands
}

// Condition 先执行 ands, 再执行 ors
type Condition struct {
	Ands Ands
	Ors  Ors
}

func (cond *Condition) Or(ands ...Ands) *Condition {
	cond.Ors = append(cond.Ors, ands...)
	return cond
}

type Ors []Ands

func (ors Ors) Or(ands Ands) Ors {
	return append(ors, ands)
}

func Where(ands Ands) *Condition {
	return &Condition{
		Ands: ands,
	}
}

package chunk

type Values interface {
	Merge(Values) Values
	Range(int, int) Values
	Len() int
}

type pseudoValues struct{}

func (p *pseudoValues) Merge(v Values) Values { return v }

func (p *pseudoValues) Range(int, int) Values { return p }

func (p *pseudoValues) Len() int { return 0 }


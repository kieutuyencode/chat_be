package pagination

type Query struct {
	Limit int `query:"limit" validate:"gte=0,lte=100"`
	Page  int `query:"page"  validate:"gte=0"`
}

func (p *Query) normalize() {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Page <= 0 {
		p.Page = 1
	}
}

func (p *Query) LimitOffset() (limit, offset int) {
	p.normalize()
	return p.Limit, (p.Page - 1) * p.Limit
}

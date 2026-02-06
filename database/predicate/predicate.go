package predicate

import "entgo.io/ent/dialect/sql"

func UnaccentContainsFold(field string, value string) func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("unaccent(").
				Ident(field).
				WriteString(") ILIKE unaccent(").
				Arg("%" + value + "%").
				WriteString(")")
		}))
	}
}

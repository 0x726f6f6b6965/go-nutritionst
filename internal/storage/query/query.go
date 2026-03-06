package query

import "github.com/Masterminds/squirrel"

type Sort struct {
	By   string
	Desc bool
}

type Query struct {
	Filters []squirrel.Sqlizer
	Sorts   []Sort
	Offset  uint64
	Limit   uint64
}

func NewQuery(filters ...squirrel.Sqlizer) *Query {
	return &Query{
		Filters: filters,
		Sorts:   []Sort{},
		Offset:  0,
		Limit:   0,
	}
}

func (q *Query) SetOffset(offset uint64) *Query {
	q.Offset = offset
	return q
}

func (q *Query) SetLimit(limit uint64) *Query {
	q.Limit = limit
	return q
}

func (q *Query) AddSortBy(by string, desc bool) *Query {
	q.Sorts = append(q.Sorts, Sort{
		By:   by,
		Desc: desc,
	})
	return q
}

func (q *Query) AddFilter(filter squirrel.Sqlizer) *Query {
	q.Filters = append(q.Filters, filter)
	return q
}

// Where builds the corresponding WHERE statement
func (q *Query) Where(b squirrel.SelectBuilder) squirrel.SelectBuilder {
	for _, f := range q.Filters {
		b = b.Where(f)
	}
	return b.PlaceholderFormat(squirrel.Dollar)
}

// Where builds the corresponding ORDER BY statement
func (q *Query) Order(b squirrel.SelectBuilder) squirrel.SelectBuilder {
	for _, s := range q.Sorts {
		if s.Desc {
			b = b.OrderBy(s.By + " DESC")
		} else {
			b = b.OrderBy(s.By)
		}
	}
	return b
}

// OffsetLimit builds the corresponding OFFSET/LIMIT statement
func (q *Query) OffsetLimit(b squirrel.SelectBuilder) squirrel.SelectBuilder {
	if q.Offset > 0 {
		b = b.Offset(q.Offset)
	}
	if q.Limit > 0 {
		b = b.Limit(q.Limit)
	}
	return b
}

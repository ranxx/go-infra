package db

import "gorm.io/gorm"

// Pager 页面
func Pager(in *gorm.DB, page, size int64) *gorm.DB {
	ret := in
	if size >= 0 {
		ret = ret.Limit(int(size))
	}
	if page > 0 && size >= 0 {
		ret = ret.Offset(int((page - 1) * size))
	}
	return ret
}

// Query query
type Query struct {
	Select        []string
	Order         interface{}
	Where         interface{}   // query
	Args          []interface{} // query args
	Offset, Limit int64
}

// SetSelect set select
func (q *Query) SetSelect(selector []string) *Query {
	q.Select = selector
	return q
}

// SetOrder set order
func (q *Query) SetOrder(order string) *Query {
	q.Order = order
	return q
}

// SetWhere set where query
func (q *Query) SetWhere(where string) *Query {
	q.Where = where
	return q
}

// SetArgs set args
func (q *Query) SetArgs(args ...interface{}) *Query {
	q.Args = args
	return q
}

// SetOffset page
func (q *Query) SetOffset(page int64) *Query {
	q.Offset = page
	return q
}

// SetLimit size
func (q *Query) SetLimit(size int64) *Query {
	q.Limit = size
	return q
}

// NewQuery new query
func NewQuery(q *Query) *Query {
	if q == nil {
		q = &Query{}
	}
	return q
}

// Q q
func (q *Query) Q(in *gorm.DB, out interface{}) (err error) {
	if q == nil {
		return
	}
	dbr := in
	if len(q.Select) > 0 {
		dbr = dbr.Select(q.Select)
	}
	if q.Where != nil && len(q.Args) > 0 {
		dbr = dbr.Where(q.Where, q.Args...)
	} else if q.Where != nil {
		dbr = dbr.Where(q.Where)
	}
	if q.Order != nil {
		dbr = dbr.Order(q.Order)
	}
	if q.Limit <= 0 {
		q.Limit = -1
	}
	err = Pager(dbr, q.Offset, q.Limit).Find(out).Error
	return
}

// One one
func (q *Query) One(in *gorm.DB, out interface{}) (err error) {
	if q == nil {
		return
	}
	dbr := in
	if len(q.Select) > 0 {
		dbr = dbr.Select(q.Select)
	}
	if q.Where != nil && len(q.Args) > 0 {
		dbr = dbr.Where(q.Where, q.Args...)
	} else if q.Where != nil {
		dbr = dbr.Where(q.Where)
	}
	if q.Order != nil {
		dbr = dbr.Order(q.Order)
	}
	err = Pager(dbr, q.Offset, q.Limit).First(out).Error
	return
}

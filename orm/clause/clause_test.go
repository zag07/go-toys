package clause

import (
	"reflect"
	"testing"
)

func TestSelect(t *testing.T) {
	var clause Clause
	clause.Set(SELECT, "User", []string{"id", "name"})
	clause.Set(WHERE, "id = ?", 1)
	clause.Set(ORDERBY, "id desc")
	clause.Set(LIMIT, 10)
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT id, name FROM User WHERE id = ? ORDER BY id desc LIMIT ?" {
		t.Error("sql failed")
	}
	if !reflect.DeepEqual(vars, []interface{}{1, 10}) {
		t.Error("vars failed")
	}
}

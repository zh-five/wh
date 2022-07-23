package wh

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type fFunc func(...interface{}) (interface{}, error)

type where struct {
	fns map[string]fFunc
}

type tagFn struct {
	fn   string
	args []string
}

var w *where

func init() {
	w = &where{
		fns: map[string]fFunc{
			"like":    fLike,
			"ftime":   fTime,
			"split":   fSplit,
			"toint":   fStr2int,
			"toint[]": fSStr2int,
		},
	}
}

func GormWhere(db *gorm.DB, param interface{}) (*gorm.DB, error) {
	return w.gormWhere(db, param)
}

func (w *where) gormWhere(db *gorm.DB, param interface{}) (*gorm.DB, error) {
	sql, vars, err := w.sqlWhere(param)
	if err != nil {
		return db, err
	}

	db = db.Where(sql, vars...)

	return db, nil
}

func (w *where) sqlWhere(param interface{}) (string, []interface{}, error) {
	pType := reflect.TypeOf(param)
	pValue := reflect.ValueOf(param)

	if pType.Kind() != reflect.Ptr || pType.Elem().Kind() != reflect.Struct {
		return "", nil, errors.New("参数应该是结构体指针")
	}

	sqls := []string{}
	vars := []interface{}{}
	for i := 0; i < pType.Elem().NumField(); i++ {
		field := pType.Elem().Field(i)

		//字段值
		tmp := pValue.Elem().FieldByName(field.Name)
		if tmp.IsZero() { //零值字段不处理
			continue
		}
		val := tmp.Interface()

		//自定义方法
		fnName := "Wh" + field.Name
		tmp = pValue.MethodByName(fnName)
		if tmp.IsValid() {
			fn, ok := tmp.Interface().(func() (string, []interface{}, error))
			if !ok {
				panic(fnName + "() Method signing error: expect func()(string, []interface{}, error)")
			}

			sql, vals, err := fn()
			if err != nil {
				if IsError(err, CodeIgnoreError) {
					continue
				}
				return "", nil, errors.WithMessage(err, fnName)
			}
			sqls = append(sqls, sql)
			vars = append(vars, vals...)
			continue
		}

		//tag 配置
		wh := field.Tag.Get("wh") // "id > ? ;split|2int"
		if wh == "" {
			continue
		}
		sql, tFns := w.parseTag(wh)

		for _, tFn := range tFns {
			args := []interface{}{val}
			for i := range tFn.args {
				args = append(args, tFn.args[i])
			}
			v, err := w.fns[tFn.fn](args...)
			if err != nil {
				return "", nil, errors.WithMessage(err, field.Name)
			}
			val = v
		}
		sqls = append(sqls, sql)
		vars = append(vars, val)
	}

	return strings.Join(sqls, " and "), vars, nil
}

//配置必须ok, 解析失败直接panic
func (w *where) parseTag(wh string) (string, []*tagFn) {
	sql := ""
	tFns := []*tagFn{}

	s := strings.SplitN(wh, ";", 2) // "name like ?;like:%?%"
	if s[0] == "" {
		panic("condition is empty: " + wh)
	}
	sql = s[0]
	if len(s) == 1 { //未配置处理函数
		return sql, tFns
	}

	ss := strings.Split(s[1], "|") // "like:%?%"
	for i := range ss {
		tmp := strings.SplitN(ss[i], ":", 2) //"time:2006-01-02"
		f := tmp[0]                          // time
		if _, ok := w.fns[f]; !ok {
			panic("tag setting error. function undefined:" + f)
		}

		as := []string{}
		if len(tmp) > 1 {
			as = strings.Split(tmp[1], ",") // ["2006-01-02"]
		}
		tFns = append(tFns, &tagFn{f, as})
	}

	return sql, tFns
}

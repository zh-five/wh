package wh

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//like  `wh:"name like ?;like:%?%"`
func fLike(v ...interface{}) (interface{}, error) {
	err := checkType(v, []reflect.Kind{
		reflect.String,
		reflect.String,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "tag function 'like'")
	}

	str := v[0].(string) // keyword
	tpl := v[1].(string) // %?%

	switch tpl {
	case "%?%":
		return "%" + str + "%", nil
	case "%?":
		return "%" + str, nil
	case "?%":
		return str + "%", nil
	default:
		return nil, errors.New("tag args error: " + tpl)
	}
}

//解析时间 `wh:"ctime > ?;ftime:2006-01-02"`
func fTime(v ...interface{}) (interface{}, error) {
	err := checkType(v, []reflect.Kind{
		reflect.String,
		reflect.String,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "tag function 'ftime'")
	}

	str := v[0].(string) //2022-07-22
	lay := v[1].(string) //2006-01-02
	t, err := time.Parse(lay, str)
	if err != nil {
		return nil, NewFormatError(err.Error())
	}

	return t, nil

}

//逗号分割字符串 `wh:"name in ?;split"`
func fSplit(v ...interface{}) (interface{}, error) {
	err := checkType(v, []reflect.Kind{
		reflect.String,
	})
	if err != nil {
		panic(err)
	}

	str := v[0].(string) //a,c

	return strings.Split(str, ","), nil
}

//string 转 int
func fStr2int(v ...interface{}) (interface{}, error) {
	if len(v) != 1 {
		panic("str2int tag seting error: too many args")
	}
	val, ok := v[0].(string)
	if !ok {
		panic("str2int tag seting error: must string")
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, IgnoreError
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return nil, NewFormatError(err.Error())
	}
	return n, nil
}

// []string --> []int
func fSStr2int(v ...interface{}) (interface{}, error) {
	if len(v) != 1 {
		panic("str2int tag seting error: too many args")
	}
	ss, ok := v[0].([]string)
	if !ok {
		panic("sstr2int tag seting error: must []string")
	}

	out := make([]int, 0, len(ss))
	for i := range ss {
		val, err := fStr2int(ss[i])
		if err != nil {
			if IsError(err, CodeIgnoreError) {
				continue
			}
		}
		out = append(out, val.(int))
	}
	if len(out) == 0 {
		return nil, IgnoreError
	}

	return out, nil
}

func checkType(v []interface{}, t []reflect.Kind) error {
	if len(v) != len(t) {
		return errors.New("args number error")
	}

	for i := range t {
		if t[i] != reflect.TypeOf(v[i]).Kind() {
			return errors.New("args type error")
		}
	}
	return nil
}

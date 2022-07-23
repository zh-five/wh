package wh

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
	//...
}

type SearchParam struct {
	IntId   int    `form:"id" wh:"id =?"`
	StrId   string `form:"id" wh:"id =?;toint"`
	Name    string `form:"name" wh:"name like ?"`
	Tags    string `form:"tag" wh:"tag in (?);split"`
	Ids     string `from:"ids" wh:"id in (?);split|toint[]"`
	Email   string `form:"name" wh:"email like ?;like:%?%"`
	MinTime string `form:"min_date" wh:"ctime >= ?;ftime:2006-01-02"`
	Month   string
}

func (s *SearchParam) WhMonth(m string) (sql string, vars []interface{}, err error) {
	//m := strings.TrimSpace(s.Month)
	if m == "" {
		err = IgnoreError
		return
	}

	t, err1 := time.Parse("2006-01", m)
	if err1 != nil {
		err = err1
		return
	}

	return "ctime >= ? and ctime < ?", []interface{}{t, t.AddDate(0, 1, 0)}, nil
}

func TestGormWhere(t *testing.T) {
	db := db()

	type args struct {
		db    *gorm.DB
		param *SearchParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "int_id",
			args: args{
				db: db,
				param: &SearchParam{
					IntId:   0,
					Name:    "name",
					Tags:    "go,rust",
					Ids:     "1,2,4",
					Email:   "email",
					MinTime: "2022-07-22",
				},
			},
			want:    "SELECT * FROM `users` WHERE name like \"name\" and tag in (\"go\",\"rust\") and id in (1,2,4) and email like \"%email%\" and ctime >= \"2022-07-22 00:00:00\"",
			wantErr: false,
		},
		{
			name: "str_id",
			args: args{
				db:    db,
				param: &SearchParam{StrId: "1"},
			},
			want:    "SELECT * FROM `users` WHERE id =1",
			wantErr: false,
		},
		{
			name: "name",
			args: args{
				db:    db,
				param: &SearchParam{Name: "abc"},
			},
			want:    "SELECT * FROM `users` WHERE name like \"abc\"",
			wantErr: false,
		},
		{
			name: "split2int",
			args: args{
				db:    db,
				param: &SearchParam{Ids: "1,2,4"},
			},
			want:    "SELECT * FROM `users` WHERE id in (1,2,4)",
			wantErr: false,
		},
		{
			name: "user_func",
			args: args{
				db:    db,
				param: &SearchParam{Month: "2022-07"},
			},
			want:    "SELECT * FROM `users` WHERE ctime >= \"2022-07-01 00:00:00\" and ctime < \"2022-08-01 00:00:00\"",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GormWhere(tt.args.db, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GormWhere() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			user := []*User{}
			stmt := got.Find(user).Statement
			sql := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
			if sql != tt.want {
				t.Errorf("sql = %v, want %v", sql, tt.want)
			}
		})
	}
}

func db() *gorm.DB {
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{DryRun: true})
	if err != nil {
		panic(err)
	}

	return db
}

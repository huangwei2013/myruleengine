package models

import (
	"github.com/astaxie/beego/logs"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
	)

type Source struct {
	Id   int64  `orm:"auto" json:"id,omitempty"`
	Name string `orm:"size(1023)" json:"name"`
	Url  string `orm:"size(1023)" json:"url"`
	CreateTime      *time.Time `json:"create_time"`
	UpdateTime      *time.Time `json:"update_time"`
}

func (*Source) TableName() string {
	return "t_source"
}

func (t *Source) Delete(id string) error {
	_, err := Ormer().Raw("DELETE FROM t_rule WHERE id = ?", id).Exec()
	return errors.Wrap(err, "database delete error")
}

func (t *Source) Update() error {
	o := orm.NewOrm()
	_, err := o.Update(t)
	if err != nil {
		logs.Error("update rule error:%v", err)
		return errors.Wrap(err, "database update error")
	}

	return errors.Wrap(err, "database insert error")
}

func (t *Source) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	if err != nil {
		logs.Error("Insert rule error:%v", err)
		return errors.Wrap(err, "database insert error")
	}

	return errors.Wrap(err, "database insert error")
}

func (*Source) Get(id string) Source {
	ts := Source{}
	qs := Ormer().QueryTable(new(Source))
	cond := orm.NewCondition()
	qs = qs.SetCond(cond.And("id__eq", id))

	qs.Limit(-1).All(&ts)
	return ts
}


func (*Source) Gets(pageNo int64, pageSize int64) []Source{
	ts := []Source{}
	qs := Ormer().QueryTable(new(Source))

	// 处理完查询条件之后
	qs.Limit(pageSize).Offset((pageNo-1)*pageSize).All(&ts)

	return ts
}

func (*Source) GetAll() []Source{
	ts := []Source{}
	qs := Ormer().QueryTable(new(Source))
	qs.Limit(-1).All(&ts)
	return ts
}
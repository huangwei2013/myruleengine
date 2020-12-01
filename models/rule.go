package models

import (
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
	)

type Rule struct {
	Id          	int64  `orm:"column(id);auto" json:"id,omitempty"`
	Expr        	string `orm:"column(expr);size(1023)" json:"expr"`
	Op          	string `orm:"column(op);size(31)" json:"op"`
	Value       	string `orm:"column(value);size(1023)" json:"value"`
	For         	string `orm:"column(for);size(1023)" json:"for"`
	SourceId      	int64  `orm:"column(source_id);" json:"source_id"`
	Summary     	string `orm:"column(summary);size(1023)" json:"summary"`
	Description 	string `orm:"column(description);size(1023)" json:"description"`
	CreateTime      *time.Time `json:"create_time"`
	UpdateTime      *time.Time `json:"update_time"`
}

func (*Rule) TableName() string {
	return "t_rule"
}

func (t *Rule) Delete(id string) error {
	_, err := Ormer().Raw("DELETE FROM t_rule WHERE id = ?", id).Exec()
	return errors.Wrap(err, "database delete error")
}

func (t *Rule) Update() error {
	o := orm.NewOrm()
	_, err := o.Update(t)
	if err != nil {
		logs.Error("update rule error:%v", err)
		return errors.Wrap(err, "database update error")
	}

	return errors.Wrap(err, "database insert error")
}

func (t *Rule) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	if err != nil {
		logs.Error("Insert rule error:%v", err)
		return errors.Wrap(err, "database insert error")
	}

	return errors.Wrap(err, "database insert error")
}

func (*Rule) Get(id string) Rule {
	ts := Rule{}
	qs := Ormer().QueryTable(new(Rule))
	cond := orm.NewCondition()
	qs = qs.SetCond(cond.And("id__eq", id))

	qs.Limit(-1).All(&ts)
	return ts
}


func (*Rule) Gets(pageNo int64, pageSize int64) []Rule{
	ts := []Rule{}
	qs := Ormer().QueryTable(new(Rule))

	// 处理完查询条件之后
	qs.Limit(pageSize).Offset((pageNo-1)*pageSize).All(&ts)

	return ts
}

func (*Rule) GetAll() []Rule{
	ts := []Rule{}
	qs := Ormer().QueryTable(new(Rule))
	qs.Limit(-1).All(&ts)
	return ts
}
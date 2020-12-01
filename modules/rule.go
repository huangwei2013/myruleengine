package modules

import (
	"gopkg.in/yaml.v2"
	"myruleengine/models"
	"strconv"
	"strings"
)

// M is map
type M map[string]interface{}

// S is slice
type S []interface{}

type Rules []models.Rule

type SourceRules struct {
	Source  models.Source
	Rules Rules
}

// RulesResp ...
type RulesResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Rules  `json:"data"`
}

// SourceResp ...
type SourceResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []models.Source `json:"data"`
}

// Content get prom rules
func (r Rules) Content() ([]byte, error) {
	rules := S{}
	for _, i := range r {
		rules = append(rules, M{
			"alert":  strconv.FormatInt(i.Id, 10),
			"expr":   strings.Join([]string{i.Expr, i.Op, i.Value}, " "),
			"for":    i.For,
			//"labels": i.Labels,
			"annotations": M{
				"rule_id":     strconv.FormatInt(i.Id, 10),
				"source_id":     strconv.FormatInt(i.SourceId, 10),
				"summary":     i.Summary,
				"description": i.Description,
			},
		})
	}
	result := M{
		"groups": S{
			M{
				"name":  "ruleengine",
				"rules": rules,
			},
		},
	}

	return yaml.Marshal(result)
}

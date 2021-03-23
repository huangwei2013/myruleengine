package modules

import (
	"context"
	"myruleengine/models"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promlog"
)

// Config ...
type Config struct {
	NotifyReties       int
	GatewayURL         string
	GatewayPathNotify  string
	EvaluationInterval model.Duration
	ReloadInterval     model.Duration
	AuthToken          string

	PromlogConfig promlog.Config
}

// Reloader ...
type Reloader struct {
	config   Config
	managers []*Manager
	logger   log.Logger
	context  context.Context
	cancel   context.CancelFunc
	running  bool
}

// NewReloader ...
func NewReloader(logger log.Logger, cfg Config) *Reloader {
	ctx, cancel := context.WithCancel(context.Background())

	reloader := Reloader{
		config: cfg,

		logger:  logger,
		context: ctx,
		cancel:  cancel,
		running: false,
	}

	return &reloader
}

// Run rule manager
func (r *Reloader) Run() {
	r.running = true
	for _, i := range r.managers {
		i.Run()
	}
}

// Stop rule manager
func (r *Reloader) Stop() {
	r.running = false
	r.cancel()
	for _, i := range r.managers {
		i.Stop()
	}

}

// download the rules and update rule manager
func (r *Reloader) Update() error {
	level.Debug(r.logger).Log("msg", "start update rule")

	sourceRules, err := r.getSourceRules()
	if err != nil {
		return err
	}

	// stop invalid manager
	for idx, m := range r.managers {
		del := true
		for _, p := range sourceRules {
			if m.Source.Id == p.Source.Id && m.Source.Url == p.Source.Url && p.Source.Url != "" {
				del = false
			}
		}
		if del {
			level.Info(r.logger).Log("msg", "source not exist, delete manager", "source_id", m.Source.Id, "source_url", m.Source.Url)
			m.Stop()
            
            if idx+1 <= len(r.managers) {
                r.managers = append(r.managers[:idx], r.managers[idx+1:]...)
            } else {
                r.managers = r.managers[:idx]
            }
		}
	}

	// update rules
	for _, p := range sourceRules {
		if p.Source.Url == "" {
			level.Error(r.logger).Log("msg", "source url is null", "source_id", p.Source.Id, "source_url", p.Source.Url)
			continue
		}
		var manager *Manager
		for _, m := range r.managers {
			if m.Source.Id == p.Source.Id && m.Source.Url == p.Source.Url && p.Source.Url != "" {
				manager = m
			}
		}
		if manager == nil {
			m, err := NewManager(r.context, r.logger, p.Source, r.config)
			if err != nil {
				level.Error(r.logger).Log("msg", "create manager error", "error", err, "source_id", manager.Source.Id, "source_url", manager.Source.Url)
				return err
			}
			m.Run()
			manager = m
			r.managers = append(r.managers, manager)
		}
		if manager != nil {
			err := manager.Update(p.Rules)
			if err != nil {
				level.Error(r.logger).Log("msg", "update rule error", "error", err, "source_id", manager.Source.Id, "source_url", manager.Source.Url)
			} else {
				level.Info(r.logger).Log("msg", "update rule success", "len", len(p.Rules), "source_id", manager.Source.Id, "source_url", manager.Source.Url)
			}
		}
	}

	level.Debug(r.logger).Log("msg", "end update rule")
	return nil
}

// Loop for checking the rules
func (r *Reloader) Loop() {
	for r.running {
		r.Update()

		select {
		case <-r.context.Done():
		case <-time.After(time.Duration(r.config.ReloadInterval)):
		}
	}
}

func (r *Reloader) getSourceRules() ([]SourceRules, error) {
	data := []SourceRules{}

	var ruleModel models.Rule
	rules := ruleModel.GetAll()
	level.Info(r.logger).Log("msg", "Number of rules ",len(rules))


	var sourceModel models.Source
	sources := sourceModel.GetAll()
	level.Info(r.logger).Log("msg", "Number of sources ",len(sources))


	data = r.genSourceRules(rules)

	//fill in url
	for idx, i := range data {
		for _, j := range sources {
			if i.Source.Id == j.Id {
				data[idx].Source.Url = j.Url
				break
			}
		}
	}

	return data, nil
}


// SourceRules cut source rules
func (r *Reloader) genSourceRules(rules []models.Rule) []SourceRules {
	tmp := map[int64]Rules{}

	for _, rule := range rules {
		if v, ok := tmp[rule.SourceId]; ok {
			tmp[rule.SourceId] = append(v, rule)
		} else {
			tmp[rule.SourceId] = Rules{rule}
		}
	}

	data := []SourceRules{}
	for id, rules := range tmp {
		data = append(data, SourceRules{
			Source:  models.Source{Id: id},
			Rules: rules,
		})
	}

	return data
}
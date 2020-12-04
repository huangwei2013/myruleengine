package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"myruleengine/models"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	client_api "github.com/prometheus/client_golang/api"
	client_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/prometheus/util/strutil"
)

// Manager ...
type Manager struct {
	Config  Config
	Source   models.Source
	Options *rules.ManagerOptions
	Manager *rules.Manager
	Rules   Rules

	logger log.Logger
}

type NotifyAlert struct {
	// Label value pairs for purpose of aggregation, matching, and disposition
	// dispatching. This must minimally include an "alertname" label.
	Labels labels.Labels `json:"labels"`

	// Extra key/value information which does not define alert identity.
	Annotations labels.Labels `json:"annotations"`

	// The known time range for this alert. Both ends are optional.
	StartsAt     time.Time `json:"startsAt,omitempty"`
	EndsAt       time.Time `json:"endsAt,omitempty"`
	GeneratorURL string    `json:"generatorURL,omitempty"`
}


// NewManager ...
func NewManager(ctx context.Context, logger log.Logger,
	source models.Source, config Config) (*Manager, error) {
	localStorage, err := NewMockStorage()
	if err != nil {
		return nil, err
	}

	options := &rules.ManagerOptions{
		Appendable: localStorage,
		TSDB:       localStorage,
		QueryFunc: HTTPQueryFunc(
			log.With(logger, "component", "http query func"),
			source.Url,
		),
		NotifyFunc: HTTPNotifyFunc(
			log.With(logger, "component", "http notify func"),
			config.AuthToken,
			fmt.Sprintf("%s%s", config.GatewayURL, config.GatewayPathNotify),
			config.NotifyReties,
			source.Url,
		),
		Context:         ctx,
		ExternalURL:     &url.URL{},
		Registerer:      nil,
		Logger:          log.With(logger, "component", "rule manager"),
		OutageTolerance: time.Hour,        // default 1h
		ForGracePeriod:  10 * time.Minute, // default 10m
		ResendDelay:     time.Minute,      // default 1m
	}
	manager := rules.NewManager(options)

	return &Manager{
		Config:  config,
		Source:    source,
		Options: options,
		Manager: manager,
		Rules:   Rules{},

		logger: logger,
	}, nil
}

// Update ...
func (m *Manager) Update(rules Rules) error {
	m.Rules = rules

	filepath := filepath.Join(os.TempDir(), fmt.Sprintf("rule.%d.yml", m.Source.Id))

	content, err := rules.Content()
	if err != nil {
		level.Error(m.logger).Log("msg", "get source rule error", "error", err, "source_id", m.Source.Id)
		return err
	}

	err = ioutil.WriteFile(filepath, content, 0644)
	if err != nil {
		level.Error(m.logger).Log("msg", "write file error", "error", err, "source_id", m.Source.Id)
		return err
	}

	return m.Manager.Update(time.Duration(m.Config.EvaluationInterval), []string{filepath})
}

// Run ...
func (m *Manager) Run() {
	level.Info(m.logger).Log("msg", "start rule manager", "source_id", m.Source.Id)
	m.Manager.Run()
}

// Stop ...
func (m *Manager) Stop() {
	level.Info(m.logger).Log("msg", "stop rule manager", "source_id", m.Source.Id)
	m.Manager.Stop()
}

// DebugNotifyFunc
func DebugNotifyFunc(logger log.Logger) rules.NotifyFunc {
	return func(ctx context.Context, expr string, alerts ...*rules.Alert) {
		for _, i := range alerts {
			level.Debug(logger).Log(
				"msg", "send alert",
				"state", i.State.String(),
				"annotations", i.Annotations.String(),
				"labels", i.Labels.String(),
			)
		}
	}
}

// Alert
type Alert rules.Alert

// MarshalJSON ...
func (a *Alert) MarshalJSON() ([]byte, error) {
	for idx, i := range a.Labels {
		if i.Name == "alertname" {
			a.Labels = append(a.Labels[:idx], a.Labels[idx+1:]...)
		}
	}
	return json.Marshal(map[string]interface{}{
		"state":        a.State,
		"labels":       a.Labels,
		"annotations":  a.Annotations,
		"value":        math.Round(a.Value*100) / 100,
		"active_at":    a.ActiveAt,
		"fired_at":     a.FiredAt,
		"resolved_at":  a.ResolvedAt,
		"last_sent_at": a.LastSentAt,
		"valid_until":  a.ValidUntil,
	})
}

// HTTPNotifyFunc
func HTTPNotifyFunc(logger log.Logger, token string, url string, retries int, sourceUrl string) rules.NotifyFunc {
	return func(ctx context.Context, expr string, alerts ...*rules.Alert) {
		if len(alerts) == 0 {
			return
		}

		var new []*NotifyAlert
		for _, alert := range alerts {
			a := &NotifyAlert{
				StartsAt:     alert.FiredAt,
				Labels:       alert.Labels,
				Annotations:  alert.Annotations,
				GeneratorURL: sourceUrl + strutil.TableLinkForExpression(expr),
			}
			if !alert.ResolvedAt.IsZero() {
				a.EndsAt = alert.ResolvedAt
			} else {
				a.EndsAt = alert.ValidUntil
			}
			new = append(new, a)
		}

		data, err := json.Marshal(new)
		if err != nil {
			level.Error(logger).Log("msg", "encode json error", "error", err, "alerts", alerts)
			return
		}
		level.Debug(logger).Log("msg", "encode alerts success", "json", data)

		for i := 1; i <= retries; i++ {
			client := http.Client{
				Timeout: 5 * time.Second, // FIXME: timeout
			}
			req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
			req.Header.Add("Token", token)
			req.Header.Add("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				level.Error(logger).Log("msg", "notify error", "url", url, "error", err, "retries", i)
				continue
			}
			if resp.StatusCode == 200 {
				level.Debug(logger).Log("msg", "notify success", "url", url)
				break
			}
			level.Error(logger).Log("msg", "notify error", "url", url, "status", resp.StatusCode, "retries", i)
		}
	}
}

// HTTPQueryFunc
// TODO: use http keep-alive
func HTTPQueryFunc(logger log.Logger, url string) rules.QueryFunc {
	client, _ := client_api.NewClient(client_api.Config{
		Address: url,
	})
	api := client_api_v1.NewAPI(client)
	return func(ctx context.Context, q string, t time.Time) (promql.Vector, error) {
		vector := promql.Vector{}

		// Query from Prometheus
		value, _, err := api.Query(ctx, q, t)
		if err != nil {
			return vector, err
		}
		switch value.Type() {
		case model.ValVector:
			for _, i := range value.(model.Vector) {
				l := labels.Labels{}
				for k, v := range i.Metric {
					l = append(l, labels.Label{
						Name:  string(k),
						Value: string(v),
					})
				}
				vector = append(vector, promql.Sample{
					Point: promql.Point{
						T: int64(i.Timestamp),
						V: float64(i.Value),
					},
					Metric: l,
				})
			}
			level.Debug(logger).Log(
				"msg", "query vector seccess",
				"query", q,
				"vector", vector,
			)
			return vector, nil
		default:
			// TODO: other type: "matrix" | "vector" | "scalar" | "string",
			return vector, fmt.Errorf("unknown result type [%s] query=[%s]", value.Type().String(), q)
		}
	}
}

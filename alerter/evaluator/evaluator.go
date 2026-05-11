package evaluator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bangmodmonitor/alerter/notifier"
	"github.com/bangmodmonitor/alerter/storage"
	"github.com/google/uuid"
)

type Evaluator struct {
	maria *storage.Maria
	ch    *storage.CH
}

func New(maria *storage.Maria, ch *storage.CH) *Evaluator {
	return &Evaluator{maria: maria, ch: ch}
}

func (e *Evaluator) Run(ctx context.Context) {
	rules, err := e.maria.GetEnabledRules(ctx)
	if err != nil {
		log.Printf("evaluator: fetch rules: %v", err)
		return
	}

	for _, rule := range rules {
		if err := e.evaluate(ctx, rule); err != nil {
			log.Printf("evaluator: rule %s (%s): %v", rule.ID, rule.Name, err)
		}
	}
}

func (e *Evaluator) evaluate(ctx context.Context, rule storage.AlertRule) error {
	var firing bool
	var details string

	switch rule.ConditionType {
	case "cpu_high":
		if rule.HostID == "" {
			return fmt.Errorf("cpu_high requires host_id")
		}
		val, err := e.ch.AvgCPU(ctx, rule.HostID, rule.DurationSec)
		if err != nil || val < 0 {
			return err
		}
		firing = val > rule.Threshold
		details = fmt.Sprintf("CPU avg %.1f%% > threshold %.1f%% over last %ds", val, rule.Threshold, rule.DurationSec)

	case "memory_high":
		if rule.HostID == "" {
			return fmt.Errorf("memory_high requires host_id")
		}
		val, err := e.ch.AvgMemory(ctx, rule.HostID, rule.DurationSec)
		if err != nil || val < 0 {
			return err
		}
		firing = val > rule.Threshold
		details = fmt.Sprintf("Memory avg %.1f%% > threshold %.1f%% over last %ds", val, rule.Threshold, rule.DurationSec)

	case "probe_down":
		if rule.TargetURL == "" {
			return fmt.Errorf("probe_down requires target_url")
		}
		stats, err := e.ch.ProbeStats(ctx, rule.TargetURL, rule.DurationSec)
		if err != nil || stats == nil {
			return err
		}
		firing = stats.DownCount == stats.Total && stats.Total > 0
		details = fmt.Sprintf("%s is DOWN (%d/%d checks failed in last %ds)",
			rule.TargetURL, stats.DownCount, stats.Total, rule.DurationSec)

	case "probe_slow":
		if rule.TargetURL == "" {
			return fmt.Errorf("probe_slow requires target_url")
		}
		stats, err := e.ch.ProbeStats(ctx, rule.TargetURL, rule.DurationSec)
		if err != nil || stats == nil {
			return err
		}
		firing = stats.AvgRespMS > rule.Threshold
		details = fmt.Sprintf("%s avg response %.0fms > threshold %.0fms over last %ds",
			rule.TargetURL, stats.AvgRespMS, rule.Threshold, rule.DurationSec)

	default:
		return fmt.Errorf("unknown condition_type: %s", rule.ConditionType)
	}

	return e.handleState(ctx, rule, firing, details)
}

func (e *Evaluator) handleState(ctx context.Context, rule storage.AlertRule, firing bool, details string) error {
	openID, err := e.maria.OpenIncident(ctx, rule.ID)
	if err != nil {
		return fmt.Errorf("open incident lookup: %w", err)
	}

	n, err := notifier.New(rule.Channel, rule.ChannelConfig)
	if err != nil {
		return fmt.Errorf("notifier: %w", err)
	}

	if firing && openID == "" {
		// New incident — fire alert
		incidentID := uuid.New().String()
		if err := e.maria.CreateIncident(ctx, incidentID, rule.ID, details); err != nil {
			return fmt.Errorf("create incident: %w", err)
		}
		subject := fmt.Sprintf("ALERT: %s", rule.Name)
		msg := fmt.Sprintf("%s\n\nTime: %s", details, time.Now().Format(time.RFC3339))
		if err := n.Send(subject, msg); err != nil {
			log.Printf("notify send (%s): %v", rule.Channel, err)
		} else {
			log.Printf("[ALERT] %s — %s", rule.Name, details)
		}
	} else if !firing && openID != "" {
		// Condition resolved
		if err := e.maria.ResolveIncident(ctx, openID); err != nil {
			return fmt.Errorf("resolve incident: %w", err)
		}
		subject := fmt.Sprintf("RESOLVED: %s", rule.Name)
		msg := fmt.Sprintf("Condition is back to normal.\n\nResolved at: %s", time.Now().Format(time.RFC3339))
		if err := n.Send(subject, msg); err != nil {
			log.Printf("notify send (%s): %v", rule.Channel, err)
		} else {
			log.Printf("[RESOLVED] %s", rule.Name)
		}
	}

	return nil
}

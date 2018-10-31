package time

import (
	"context"
	"github.com/marvin-automator/marvin/actions"
	"github.com/gorhill/cronexpr"
	"time"
)

func init() {
	p := actions.Registry.AddProvider("Time", "Time-related actions", []byte{})
	g := p.AddGroup("cron", "Cron-related tasks", []byte{})
	g.AddManualTrigger("Cron", "Schedule a function to run on an interval based on a Cron expression.", []byte{}, cronTrigger)
}

type CronInput struct {
	Expression string
}

type CronEvent struct {
	Time time.Time `json:"time"`
}

func cronTrigger(in CronInput, ctx context.Context) (<-chan CronEvent, error) {
	expr, err := cronexpr.Parse(in.Expression)
	if err != nil {
		return nil, err
	}

	out := make(chan CronEvent)
	var f func()
	f = func() {
		now := time.Now()
		n := expr.Next(now)
		t := time.NewTimer(n.Sub(now))

		select {
		case <-ctx.Done():
		case now = <-t.C:
			out <- CronEvent{now}
			f()
		}
	}

	go f()
	return out, nil
}

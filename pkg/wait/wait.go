package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Wait interface {
	Wait(ctx context.Context) bool
}

type schedule struct {
	cron *cron.Cron
	tick chan struct{}
}

func (s *schedule) Wait(ctx context.Context) bool {
	s.cron.Start()
	select {
	case <-ctx.Done():
		s.cron.Stop()
		return false
	case <-s.tick:
		return true
	}
}

func NewSchedule(expression string) (Wait, error) {
	ticker := make(chan struct{})
	logger := cron.VerbosePrintfLogger(logrus.StandardLogger().WithField("name", "refresh-schedule-cron"))
	job := cron.SkipIfStillRunning(logger)(cron.FuncJob(func() { ticker <- struct{}{} }))
	c := cron.New(cron.WithLogger(logger))
	if _, err := c.AddJob(expression, job); err != nil {
		return nil, fmt.Errorf("failed to parse cron expression %q: %w", expression, err)
	}
	return &schedule{cron: c, tick: ticker}, nil
}

type interval struct {
	interval time.Duration
}

func (i *interval) Wait(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(i.interval):
		return true
	}
}

func NewInterval(duration string) (Wait, error) {
	intval, err := time.ParseDuration(duration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration %q: %w", duration, err)
	}
	return &interval{interval: intval}, nil
}

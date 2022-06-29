package cron

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws"
	"sort"
	"time"
)

type EventCron struct {
	*Cron
}

func (c *EventCron) Init() (action gnet.Action) {
	now := c.Cron.now()
	for _, entry := range c.Cron.entries {
		entry.Next = entry.Schedule.Next(now)
		c.Cron.logger.Info("schedule", "now", now, "entry", entry.ID, "next", entry.Next)
	}

	return action
}

func (c *EventCron) Ticker() (delay time.Duration, action gnet.Action) {
	now := c.Cron.now()

	// Run every entry whose next time was less than now
	for _, e := range c.Cron.entries {
		if e.Next.After(now) || e.Next.IsZero() {
			break
		}

		c.startJob(e.WrappedJob)

		e.Prev = e.Next
		e.Next = e.Schedule.Next(now)
		c.Cron.logger.Info("run", "now", now, "entry", e.ID, "next", e.Next)
	}

	// Determine the next entry to run.
	sort.Sort(byTime(c.Cron.entries))

	if len(c.Cron.entries) == 0 || c.Cron.entries[0].Next.IsZero() {
		// If there are no entries yet, just sleep - it still handles new entries
		// and stop requests.
		return 100000 * time.Hour, action
	} else {
		return c.Cron.entries[0].Next.Sub(now), action
	}
}

func (c *EventCron) startJob(j Job) {
	c.Cron.jobWaiter.Add(1)
	if err := gnetws.GoroutinePool().Submit(func() {
		defer c.Cron.jobWaiter.Done()
		j.Run()
	}); err != nil {
		c.logger.Error(err, "startJob error")
	}
}

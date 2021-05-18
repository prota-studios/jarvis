package calendar

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar struct {
	calendar *calendar.Service
}

func NewCalendar(credsFile string) (c *Calendar, err error) {
	c = &Calendar{}
	ctx := context.Background()

	c.calendar, err = calendar.NewService(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return
	}
	go func() {
		err := c.GetWeekly()
		if err != nil {
			logrus.Error(err)
		}
	}()
	return
}

func (c *Calendar) GetWeekly() (err error) {
	//var event *calendar.Event
	//
	var events *calendar.Events
	events, err = c.calendar.Events.List("c29sLmNhdGVzQHByb3Rhc3R1ZGlvcy5jb20").Do()
	////event, err = c.calendar.Events.Get("c29sLmNhdGVzQHByb3Rhc3R1ZGlvcy5jb20", "NjU3YzIzNGVlb3A1MWpzczFmanJ2NzRiMThfMjAyMTA1MThUMjAxNTAwWiBzb2wuY2F0ZXNAcHJvdGFzdHVkaW9zLmNvbQ").Do()
	if err != nil {
		return err
	}
	spew.Dump(events)

	return
}

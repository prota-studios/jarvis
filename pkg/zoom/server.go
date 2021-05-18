package zoom

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/himalayan-institute/zoom-lib-golang"
	"github.com/prota-studios/jarvis/models"
	"github.com/prota-studios/jarvis/pkg/humanize"
	"github.com/prota-studios/jarvis/restapi/operations"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultZoomDelay = 10

type Config struct {
	ZoomAPIKey    string        `long:"zoom-api-key" description:"Zoom API Key" required:"false" default:"" env:"ZOOM_API_KEY"`
	ZoomAPISecret string        `long:"zoom-api-secret" description:"Zoom API Secret" required:"false" default:"" env:"ZOOM_API_SECRET"`
	ZoomUserId    string        `long:"zoom-user-id" description:"Zoom User ID" required:"false" default:"" env:"ZOOM_USER_ID"`
	ZoomInterval  time.Duration `long:"zoom-interval" description:"Zoom Ticker Interval" required:"false" default:"5m" env:"ZOOM_INTERVAL"`
}
type Server struct {
	config *Config
	client *zoom.Client
}

func NewServer(cfg *Config) *Server {
	z := &Server{config: cfg}
	if z.config.ZoomInterval == 0 {
		z.config.ZoomInterval = defaultZoomDelay
	}
	z.client = zoom.NewClient(cfg.ZoomAPIKey, cfg.ZoomAPISecret)

	return z
}

func (z *Server) ListUpcomingMeetings() middleware.Responder {

	resp := operations.NewListUpcomingMeetingsOK()
	// Prepare response
	body := &operations.ListUpcomingMeetingsOKBody{
		Meetings: z.listUpcomingMeetings(),
	}

	resp.WithPayload(body)
	return resp
}

func (z *Server) listUpcomingMeetings() (meetings []*models.Meeting) {
	opts := zoom.ListMeetingsOptions{
		Type:   zoom.ListMeetingTypeUpcoming,
		HostID: z.config.ZoomUserId,
	}
	zMeetings, err := z.client.ListMeetings(opts)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	for _, meeting := range zMeetings.Meetings {
		m := &models.Meeting{
			ID:   swag.Int64(int64(meeting.ID)),
			Name: swag.String(meeting.Topic),
		}
		logrus.Infof("%+v\\n", meeting)
		meetings = append(meetings, m)
	}

	return
}

func (z *Server) Start() {
	ctx, done := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	// goroutine to check for signals to gracefully finish all functions
	g.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-signalChannel:
			logrus.Infof("Received signal: %s", sig)
			done()
		case <-gctx.Done():
			fmt.Printf("closing signal goroutine")
			return gctx.Err()
		}

		return nil
	})

	// just a ticker every 2s
	g.Go(func() error {
		// Initial update
		z.listUpcomingMeetings()
		ticker := time.NewTicker(z.config.ZoomInterval)
		for {
			select {
			case <-ticker.C:
				z.listUpcomingMeetings()
				logrus.Infof("%s zoom ticks", humanize.HumanizeDuration(z.config.ZoomInterval),
				)

				// testcase what happens if an error occured
				//return fmt.Errorf("test error ticker 2s")
			case <-gctx.Done():
				return gctx.Err()
			}
		}
	})

	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logrus.Infof("context was canceled")
		} else {
			logrus.Errorf("received error: %v", err)
		}
	} else {
		logrus.Info("finished clean")
	}
	return
}

package zoom

import (
	"context"
	"errors"
	"fmt"
	"github.com/prota-studios/javis/models"
	"github.com/prota-studios/javis/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/himalayan-institute/zoom-lib-golang"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultZoomDelay = 10

type Config struct {
	ZoomAPIKey    string        `long:"zoom-api-key" description:"Zoom API Key" required:"true" default:"" env:"ZOOM_API_KEY"`
	ZoomAPISecret string        `long:"zoom-api-secret" description:"Zoom API Secret" required:"true" default:"" env:"ZOOM_API_SECRET"`
	ZoomUserId    string        `long:"zoom-user-id" description:"Zoom User ID" required:"true" default:"9meQ5bpuTHOXXKs6I8Hw0A" env:"ZOOM_USER_ID"`
	ZoomDelay     time.Duration `long:"zoom-delay" description:"Zoom Ticker Delay" required:"true" default:"10s" env:"ZOOM_DELAY"`
	DBType        string        `long:"db-type" description:"Database type" required:"true" default:"sqlite" env:"DB_TYPE"`
	DBUrl         string        `long:"db-url" description:"Database connection url" required:"true" default:"db.sqlite" env:"DB_URL"`
	Debug         bool          `long:"debug" description:"Debug logging" required:"false"  env:"DEBUG"`
}
type Zoom struct {
	config *Config
	client *zoom.Client
}

func NewZoom(cfg *Config) *Zoom {
	z := &Zoom{config: cfg}
	if z.config.ZoomDelay == 0 {
		z.config.ZoomDelay = defaultZoomDelay
	}
	z.client = zoom.NewClient(cfg.ZoomAPIKey, cfg.ZoomAPISecret)

	return z
}

func (z *Zoom) ListUpcomingMeetings() middleware.Responder {

	resp := operations.NewListUpcomingMeetingsOK()
	// Prepare response
	body := &operations.ListUpcomingMeetingsOKBody{
		Meetings: z.listUpcomingMeetings(),
	}

	resp.WithPayload(body)
	return resp
}
func (z *Zoom) listUpcomingMeetings() (meetings []*models.Meeting) {
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
			ID:   int64(meeting.ID),
			Name: meeting.Topic,
		}
		logrus.Infof("%+v\\n", meeting)
		meetings = append(meetings, m)
	}

	return
}

func (z *Zoom) Start() {
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
		ticker := time.NewTicker(z.config.ZoomDelay)
		for {
			select {
			case <-ticker.C:
				z.listUpcomingMeetings()
				logrus.Infof("%fs ticked, refreshing upcoming meetings", z.config.ZoomDelay.Seconds())

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

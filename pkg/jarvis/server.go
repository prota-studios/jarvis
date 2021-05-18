package jarvis

/*
  #include <stdio.h>
  #include <unistd.h>
  #include <termios.h>
  char getch(){
      char ch = 0;
      struct termios old = {0};
      fflush(stdout);
      if( tcgetattr(0, &old) < 0 ) perror("tcsetattr()");
      old.c_lflag &= ~ICANON;
      old.c_lflag &= ~ECHO;
      old.c_cc[VMIN] = 1;
      old.c_cc[VTIME] = 0;
      if( tcsetattr(0, TCSANOW, &old) < 0 ) perror("tcsetattr ICANON");
      if( read(0, &ch,1) < 0 ) perror("read()");
      old.c_lflag |= ICANON;
      old.c_lflag |= ECHO;
      if(tcsetattr(0, TCSADRAIN, &old) < 0) perror("tcsetattr ~ICANON");
      return ch;
  }
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/gordonklaus/portaudio"
	"github.com/prota-studios/jarvis/models"
	"github.com/prota-studios/jarvis/pkg/calendar"
	"github.com/prota-studios/jarvis/pkg/zoom"
	"github.com/prota-studios/jarvis/restapi/operations"
	"github.com/shomali11/slacker"
	"github.com/sirupsen/logrus"
	wave "github.com/zenwerk/go-wave"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	ZoomConfig    *zoom.Config
	ZoomEnabled   bool          `long:"zoom-enable" description:"enable zoom integration" required:"false"  env:"ZOOM_ENABLED"`
	Interval      time.Duration `long:"interval" description:"Server Interval" required:"true" default:"5s" env:"INTERVAL"`
	DBType        string        `long:"db-type" description:"Database type" required:"true" default:"sqlite" env:"DB_TYPE"`
	DBUrl         string        `long:"db-url" description:"Database connection url" required:"true" default:"db.sqlite" env:"DB_URL"`
	Debug         bool          `long:"debug" description:"Debug logging" required:"false"  env:"DEBUG"`
	SlackAPIToken string        `long:"slack-api-token" description:"Slack API Token key" required:"false" env:"SLACK_API_TOKEN"`
	CredsFile     string        `long:"creds-file" description:"Credentials for Google application access" required:"false" default:"creds.json" env:"CREDS_FILE"`
}

type Server struct {
	config  *Config
	zoom    *zoom.Server
	ip      string
	slacker *slacker.Slacker

	recording    bool
	recordingSig chan os.Signal

	calendar *calendar.Calendar
}

func NewServer(config *Config) (s *Server, err error) {
	s = &Server{config: config}
	s.slacker = slacker.NewClient(config.SlackAPIToken)
	s.calendar, err  = calendar.NewCalendar(s.config.CredsFile)
	return
}

func (s *Server) WhatMyIpAddress() {

	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	s.ip = string(ip)
}

func (s *Server) Start() {
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

	// Check where we are
	g.Go(func() error {
		// Initial update
		ticker := time.NewTicker(s.config.Interval)
		for {
			select {
			case <-ticker.C:
				s.WhatMyIpAddress()


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

func (s *Server) DictationStatus() middleware.Responder {
	resp := operations.NewDictationStatusOK()
	ds := &models.DictationStatus{
		Processing: swag.Bool(false),
		Recording:  swag.Bool(false),
	}
	resp.SetPayload(ds)
	return resp
}

func (s *Server) StartDictation() middleware.Responder {

	if !s.recording {
		go func() {
			err := s.recordDictation()
			if err != nil {
				logrus.Errorf("error with recording: %v", err)
			}
		}()
	}

	resp := operations.NewStartOK()
	ds := &models.DictationStatus{
		Processing: swag.Bool(false),
		Recording:  swag.Bool(true),
	}
	resp.SetPayload(ds)

	return resp
}

type Dictation struct {
}

func (s *Server) recordDictation() (err error) {
	s.recording = true
	s.recordingSig = make(chan os.Signal, 1)
	signal.Notify(s.recordingSig, os.Interrupt, os.Kill)

	var f *os.File
	var stream *portaudio.Stream

	f, err = ioutil.TempFile(os.TempDir(), "dictation-*.wav")
	if err != nil {
		return err
	}
	os.RemoveAll("test.wav")
	f, err = os.Create("test.wav")
	if f == nil {
		logrus.Fatal("Can't open wave file")
	}

	var writer *wave.Writer
	writer, err = wave.NewWriter(wave.WriterParam{
		Out:           f,
		Channel:       1,
		SampleRate:    44100,
		BitsPerSample: 8,
	})
	//writer := wave.NewWriter(f, 2, 2, 44100, 16)
	defer func() {
		key := C.getch()
		logrus.Info("Cleaning up ...")
		if key == 27 {
			// better to control
			// how we close then relying on defer
			writer.Close()
			stream.Close()
			portaudio.Terminate()
			os.Exit(0)

		}

	}()
	err = portaudio.Initialize()
	if err != nil {
		return err
	}

	in := make([]byte, 64)

	if stream, err = portaudio.OpenDefaultStream(1, 0, float64(44100), len(in), in); err != nil {
		return
	}
	defer stream.Close()
	if err = stream.Start(); err != nil {
		return
	}

	logrus.Infof("Beginning recording to %s", f.Name())
loop:
	for {
		err = stream.Read()
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(in))
		if err != nil {
			return err
		}
		//err = binary.Write(f, binary.BigEndian, in)
		//if err != nil {
		//	return err
		//}
		select {
		case <-s.recordingSig:
			break loop
		default:
		}
	}
	logrus.Infof("Finishing up the recording to %s", f.Name())

	s.recording = false
	err = stream.Stop()
	if err != nil {
		return err
	}
	return
}
func (s *Server) StopDictation() middleware.Responder {

	if s.recording {
		s.recordingSig <- os.Interrupt
	}

	resp := operations.NewStopOK()
	ds := &models.DictationStatus{
		Processing: swag.Bool(true),
		Recording:  swag.Bool(false),
	}
	resp.SetPayload(ds)

	return resp
}

// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/prota-studios/jarvis/pkg/jarvis"
	"github.com/prota-studios/jarvis/pkg/zoom"
	"github.com/prota-studios/jarvis/restapi/operations"
	"github.com/sirupsen/logrus"
	"net/http"
)

//go:generate swagger generate server --target ../../jarvis --name Jarvis --spec ../swagger.yaml --principal interface{}

func configureFlags(api *operations.JarvisAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
	opts := &jarvis.Config{}
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		{
			ShortDescription: "Config options",
			LongDescription:  "Configurable options for tod server",
			Options:          opts,
		},
	}
}

func configureAPI(api *operations.JarvisAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	cfg := api.CommandLineOptionsGroups[0].Options.(*jarvis.Config)

	// Turn on teh Server Bot....
	j, err := jarvis.NewServer(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	go j.Start()

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	//api.UseRedoc()
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	api.Logger = logrus.Infof
	if cfg.Debug {
		api.Logger = logrus.Debugf
	}

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Health
	api.HealthHandler = operations.HealthHandlerFunc(func(params operations.HealthParams) middleware.Responder {
		return operations.NewHealthOK()
	})

	// Dictation
	api.DictationStatusHandler = operations.DictationStatusHandlerFunc(func(params operations.DictationStatusParams) middleware.Responder {

		return j.DictationStatus()
	})
	api.StartHandler = operations.StartHandlerFunc(func(params operations.StartParams) middleware.Responder {

		return j.StartDictation()
	})

	api.StopHandler = operations.StopHandlerFunc(func(params operations.StopParams) middleware.Responder {

		return j.StopDictation()
	})

	// If Zoom is enabled, handle all those operations.
	if cfg.ZoomEnabled {
		z := zoom.NewServer(cfg.ZoomConfig)
		go z.Start()
		api.GetMeetingHandler = operations.GetMeetingHandlerFunc(func(params operations.GetMeetingParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetMeeting has not yet been implemented")
		})
		api.ListUpcomingMeetingsHandler = operations.ListUpcomingMeetingsHandlerFunc(func(params operations.ListUpcomingMeetingsParams) middleware.Responder {
			return z.ListUpcomingMeetings()
		})
		api.ListRecordingsHandler = operations.ListRecordingsHandlerFunc(func(params operations.ListRecordingsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.ListRecordings has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}

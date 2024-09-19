package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	app "lib/app"
	cache "lib/cache"
	config "lib/config"
	extractor "lib/extractor"
	logger "lib/logger"
	streamer "lib/streamer"
)

const appVersion = "2.0.0"

type flagsT struct {
	version bool
	config  string
}

const (
	NoError = iota
	ConfigError
	LoggerError
	ExtractorError
	StreamerError
	WebServerError
	CacheError
)

func parseCLIFlags() flagsT {
	var f flagsT
	flag.BoolVar(&f.version, "version", false, "prints current yt-proxy version")
	flag.StringVar(&f.config, "config", "config.jsonc", "config file path")
	flag.Parse()
	return f
}

func main() {
	flags := parseCLIFlags()
	if flags.version {
		stdout(appVersion)
		os.Exit(NoError)
	}
	startApp(flags.config)
}

var stdout = func(s string) { os.Stdout.WriteString(fmt.Sprintf("%s\n", s)) }
var stderr = func(s string) { os.Stderr.WriteString(fmt.Sprintf("ERROR   %s\n", s)) }
var checkOrExit = func(err error, name string, errorcode int) {
	if err != nil {
		stderr(fmt.Sprintf("%s error. %s", name, err))
		os.Exit(errorcode)
	}
}
var texts = [5]string{
	"Config",
	"Logger",
	"Extractor",
	"Cache",
	"Streamer",
}

func startApp(conf_file string) {
	conf, err := config.Read(conf_file)
	checkOrExit(err, texts[0], ConfigError)

	log, err := logger.New(conf.Log)
	checkOrExit(err, texts[1], LoggerError)

	opts := make([]app.Option, 0)
	for _, v := range conf.SubConfig {
		opt := getNewApp(log, v)
		opts = append(opts, opt)
	}
	defapp := getNewApp(log, config.SubConfigT{
		ConfigT: config.ConfigT{
			Streamer:  conf.Streamer,
			Extractor: conf.Extractor,
			Cache:     conf.Cache,
		},
		Name: "default",
	})
	app := app.New(
		logger.NewLayer(log, "App"),
		defapp,
		opts)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Bad request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("Bad request", r)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Play request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("User request", r)
		app.Run(w, r)
		log.LogInfo("Player disconnected", r.RemoteAddr)
	})
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%d", conf.Host, conf.PortInt),
	}
	go signalsCatcher(log, app, s)

	log.LogInfo("Starting web server", "host", conf.Host, "port", conf.PortInt)
	if err = s.ListenAndServe(); err == http.ErrServerClosed {
		log.LogInfo("HTTP server closed")
		os.Exit(NoError)
	} else {
		log.LogError("HTTP server", err)
		log.Close()
		os.Exit(WebServerError)
	}
}

func signalsCatcher(log logger.T, app *app.T, s *http.Server) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		switch <-sigint {
		case syscall.SIGHUP:
			log.LogWarning("Config reloading")
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			log.LogWarning("Exiting")
			app.Shutdown()
			if err := s.Shutdown(context.Background()); err != nil {
				// Error from closing listeners, or context timeout:
				log.LogError("HTTP server Shutdown", err)
			}
			return
		}
	}
}

func getNewApp(log logger.T, v config.SubConfigT) app.Option {
	subcheck := func(err error, name string, errorcode int) {
		checkOrExit(err, v.Name+" "+name, errorcode)
	}
	newname := func(name string) string {
		return fmt.Sprintf("[%s] %s", v.Name, name)
	}
	xtr, err := extractor.New(v.Extractor, logger.NewLayer(log, newname(texts[2])))
	subcheck(err, texts[2], ExtractorError)
	cch, err := cache.New(v.Cache, logger.NewLayer(log, newname(texts[3])))
	subcheck(err, texts[3], CacheError)
	strm, err := streamer.New(v.Streamer, logger.NewLayer(log, newname(texts[4])), xtr)
	subcheck(err, texts[4], StreamerError)
	return app.Option{
		Name:  v.Name,
		Sites: v.Sites,
		X:     xtr,
		S:     strm,
		C:     cch,
		L:     logger.NewLayer(log, fmt.Sprintf("[%s] app", v.Name)),
	}
}

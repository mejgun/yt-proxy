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
	SomeError
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
		os.Stdout.WriteString(fmt.Sprintf("%s\n", appVersion))
		os.Exit(NoError)
	}
	startApp(flags.config)
}

func startApp(conf_file string) {
	conf, def, opts, log, err := readConfig(conf_file)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Config read error: %s\n", err))
		os.Exit(SomeError)
	}
	app := app.New(
		logger.NewLayer(log, "App"),
		def, opts)

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
	shouldWait := make(chan confChan)
	go signalsCatcher(conf_file, log, app, shouldWait)
	httpLoop(log, conf, shouldWait)
}

func makeServer(conf config.ConfigT) *http.Server {
	return &http.Server{Addr: fmt.Sprintf("%s:%d", conf.Host, conf.PortInt)}
}

type confChan struct {
	cnf     config.ConfigT
	restart bool
}

func httpLoop(log logger.T, conf config.ConfigT, ch <-chan confChan) {
	for {
		log.LogInfo("Starting web server", "host", conf.Host, "port", conf.PortInt)
		s := makeServer(conf)
		done := make(chan struct{})
		go startHttp(s, log, done)
		<-ch
		log.LogInfo("Stopping web server")
		s.Close()
		s.Shutdown(context.Background())
		<-done
		msg := <-ch
		if !msg.restart {
			log.LogInfo("Web server stopped")
			os.Exit(NoError)
		}
		conf = msg.cnf
	}
}

func startHttp(s *http.Server, log logger.T, done chan<- struct{}) {
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.LogError("HTTP server error", err)
		log.Close()
		os.Exit(SomeError)
	}
	done <- struct{}{}
}

func readConfig(conf_file string) (config.ConfigT, app.Option, []app.Option,
	logger.T, error) {
	conf, err := config.Read(conf_file)
	if err != nil {
		return config.ConfigT{}, app.Option{}, nil, nil, err
	}

	log, err := logger.New(conf.Log)
	if err != nil {
		return config.ConfigT{}, app.Option{}, nil, nil, err
	}

	defapp, err := getNewApp(log, config.SubConfigT{
		ConfigT: config.ConfigT{
			Streamer:  conf.Streamer,
			Extractor: conf.Extractor,
			Cache:     conf.Cache,
		},
		Name: "default",
	})
	if err != nil {
		return config.ConfigT{}, app.Option{}, nil, nil, err
	}

	opts := make([]app.Option, 0)
	for _, v := range conf.SubConfig {
		opt, err := getNewApp(log, v)
		if err != nil {
			return config.ConfigT{}, app.Option{}, nil, nil, err
		}
		opts = append(opts, opt)
	}
	return conf, defapp, opts, log, nil
}

func signalsCatcher(conf_file string, log logger.T, app *app.AppLogic,
	ch chan<- confChan) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)
	for {
		switch <-sigint {
		case syscall.SIGHUP:
			log.LogWarning("Config reloading")
			conf, def, opts, lognew, err := readConfig(conf_file)
			if err != nil {
				log.LogError("Config reload error", err)
			} else {
				ch <- confChan{}
				app.ReloadConfig(logger.NewLayer(lognew, "App"), def, opts)
				log = lognew
				ch <- confChan{conf, true}
			}
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			log.LogWarning("Exiting")
			ch <- confChan{}
			app.Shutdown()
			ch <- confChan{}
		}
	}
}

func getNewApp(log logger.T, v config.SubConfigT) (app.Option, error) {
	texts := [3]string{
		"Extractor",
		"Cache",
		"Streamer",
	}

	newname := func(name string) string {
		return fmt.Sprintf("[%s] %s", v.Name, name)
	}
	nameerr := func(name string, err error) error {
		return fmt.Errorf("%s: %s", newname(name), err)
	}
	xtr, err := extractor.New(v.Extractor,
		logger.NewLayer(log, newname(texts[0])))
	if err != nil {
		return app.Option{}, nameerr(texts[0], err)
	}
	cch, err := cache.New(v.Cache,
		logger.NewLayer(log, newname(texts[1])))
	if err != nil {
		return app.Option{}, nameerr(texts[1], err)
	}
	strm, err := streamer.New(v.Streamer,
		logger.NewLayer(log, newname(texts[2])), xtr)
	if err != nil {
		return app.Option{}, nameerr(texts[2], err)
	}
	return app.Option{
		Name:               v.Name,
		Sites:              v.Sites,
		X:                  xtr,
		S:                  strm,
		C:                  cch,
		L:                  logger.NewLayer(log, fmt.Sprintf("[%s] app", v.Name)),
		DefaultVideoHeight: v.DefaultVideoHeight,
	}, nil
}

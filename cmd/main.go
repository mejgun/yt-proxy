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
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%d", conf.Host, conf.PortInt),
	}
	go signalsCatcher(conf_file, log, app, s)

	log.LogInfo("Starting web server", "host", conf.Host, "port", conf.PortInt)
	if err = s.ListenAndServe(); err == http.ErrServerClosed {
		log.LogInfo("HTTP server closed")
		os.Exit(NoError)
	} else {
		log.LogError("HTTP server", err)
		log.Close()
		os.Exit(SomeError)
	}
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

func signalsCatcher(conf_file string, log logger.T, app *app.AppLogic, s *http.Server) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		switch <-sigint {
		case syscall.SIGHUP:
			log.LogWarning("Config reloading")
			_, def, opts, lognew, err := readConfig(conf_file)
			if err != nil {
				log.LogError("Config reload error", err)
			} else {
				app.ReloadConfig(logger.NewLayer(lognew, "App"), def, opts)
				log = lognew
			}
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
		Name:  v.Name,
		Sites: v.Sites,
		X:     xtr,
		S:     strm,
		C:     cch,
		L:     logger.NewLayer(log, fmt.Sprintf("[%s] app", v.Name)),
	}, nil
}

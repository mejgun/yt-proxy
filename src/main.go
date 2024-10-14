// Package main is entry point
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	cache_mux "ytproxy/cache/mux"
	config "ytproxy/config"
	extractor "ytproxy/extractor"
	logger "ytproxy/logger"
	logic "ytproxy/logic"
	streamer "ytproxy/streamer"
	"ytproxy/utils"
)

const appVersion = "2.3.0"

type flagsT struct {
	version bool
	config  string
}

const (
	noError = iota
	someError
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
		_, _ = os.Stdout.WriteString(fmt.Sprintf("%s\n", appVersion))
		os.Exit(noError)
	}
	startApp(flags.config)
}

func startApp(confFile string) {
	conf, def, opts, log, err := readConfig(confFile)
	if err != nil {
		utils.WriteError(fmt.Errorf("config read error: %s", err))
		os.Exit(someError)
	}
	appLogic := logic.New(def, opts)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Bad request", "addr", r.RemoteAddr, "url", r.RequestURI)
		log.LogDebug("Bad request", "req", r)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		appLogic.Run(w, r, log)
	})
	shouldWait := make(chan confChan)
	go signalsCatcher(confFile, log, appLogic, shouldWait)
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
		go startHTTP(s, log, done)
		<-ch
		log.LogInfo("Stopping web server")
		if err := s.Close(); err != nil {
			log.LogInfo("Web server stopping", "error", err)
		}
		if err := s.Shutdown(context.Background()); err != nil {
			log.LogInfo("Web server shutting down", "error", err)
		}
		<-done
		msg := <-ch
		if !msg.restart {
			log.LogInfo("Web server stopped")
			os.Exit(noError)
		}
		conf = msg.cnf
	}
}

func startHTTP(s *http.Server, log logger.T, done chan<- struct{}) {
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.LogError("HTTP server", "error", err)
		log.Close()
		os.Exit(someError)
	}
	done <- struct{}{}
}

func readConfig(confFile string) (config.ConfigT, logic.Option, []logic.Option,
	logger.T, error) {
	conf, err := config.Read(confFile)
	if err != nil {
		return config.ConfigT{}, logic.Option{}, nil, nil, err
	}

	log, err := logger.New(conf.Log)
	if err != nil {
		return config.ConfigT{}, logic.Option{}, nil, nil, err
	}

	defaultAppLogic, err := getNewAppLogic(log, config.SubConfigT{
		ConfigT: config.ConfigT{
			Streamer:           conf.Streamer,
			Extractor:          conf.Extractor,
			Cache:              conf.Cache,
			DefaultVideoHeight: conf.DefaultVideoHeight,
			MaxVideoHeight:     conf.MaxVideoHeight,
			Sites:              conf.Sites,
		},
		Name: "default",
	})
	if err != nil {
		return config.ConfigT{}, logic.Option{}, nil, nil, err
	}

	optionalAppLogic := make([]logic.Option, 0)
	for _, v := range conf.SubConfig {
		opt, err := getNewAppLogic(log, v)
		if err != nil {
			return config.ConfigT{}, logic.Option{}, nil, nil, err
		}
		optionalAppLogic = append(optionalAppLogic, opt)
	}
	return conf, defaultAppLogic, optionalAppLogic, log, nil
}

func signalsCatcher(confFile string, log logger.T, appLogic *logic.AppLogic,
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
			conf, def, opts, logNew, err := readConfig(confFile)
			if err != nil {
				log.LogError("Config reload", "error", err)
			} else {
				ch <- confChan{}
				appLogic.ReloadConfig(logNew, def, opts)
				log = logNew
				ch <- confChan{conf, true}
			}
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			log.LogWarning("Exiting")
			ch <- confChan{}
			appLogic.Shutdown(log)
			ch <- confChan{}
		}
	}
}

func getNewAppLogic(log logger.T, v config.SubConfigT) (logic.Option, error) {
	texts := [3]string{
		"Extractor",
		"Cache",
		"Streamer",
	}

	newName := func(name string) string {
		return fmt.Sprintf("[%s] %s", v.Name, name)
	}
	nameErr := func(name string, err error) error {
		return fmt.Errorf("%s: %s", newName(name), err)
	}
	xtr, err := extractor.New(v.Extractor,
		logger.NewLayer(log, newName(texts[0])))
	if err != nil {
		return logic.Option{}, nameErr(texts[0], err)
	}
	cch, err := cache_mux.New(v.Cache,
		logger.NewLayer(log, newName(texts[1])))
	if err != nil {
		return logic.Option{}, nameErr(texts[1], err)
	}
	strm, err := streamer.New(v.Streamer,
		logger.NewLayer(log, newName(texts[2])), xtr)
	if err != nil {
		return logic.Option{}, nameErr(texts[2], err)
	}
	return logic.Option{
		Name:               v.Name,
		Sites:              v.Sites,
		X:                  xtr,
		S:                  strm,
		C:                  cch,
		DefaultVideoHeight: v.DefaultVideoHeight,
		MaxVideoHeight:     v.MaxVideoHeight,
	}, nil
}

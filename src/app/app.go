// Package app implements apps real entry point
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	cache_mux "ytproxy/cache/mux"
	config "ytproxy/config"
	extractor_mux "ytproxy/extractor/mux"
	logger "ytproxy/logger"
	logger_mux "ytproxy/logger/mux"
	logic "ytproxy/logic"
	streamer "ytproxy/streamer"
)

// Run creates and runs all objects
func Run(confFile string) error {
	conf, def, opts, log, err := readConfig(confFile)
	if err != nil {
		return fmt.Errorf("config read error: %s", err)

	}
	appLogic := logic.New(def, opts)
	shouldWait := make(chan confChan)
	go signalsCatcher(confFile, log, shouldWait)
	return httpLoop(log, conf, appLogic, shouldWait)
}

func makeServer(conf config.T, log logger.T,
	appLogic *logic.AppLogic) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf("%s:%d", conf.Host, conf.PortInt),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.RequestURI, "/play/"):
				appLogic.Run(w, r, log)
			default:
				log.LogInfo("Bad request", "addr", r.RemoteAddr, "url", r.RequestURI)
				log.LogDebug("Bad request", "req", r)
				http.NotFound(w, r)
			}
		})}
}

type confChan struct {
	cnf      config.T
	appLogic *logic.AppLogic
	restart  bool
}

func httpLoop(log logger.T, conf config.T, appLogic *logic.AppLogic,
	ch <-chan confChan) error {
	for {
		log.LogInfo("Starting web server", "host", conf.Host, "port", conf.PortInt)
		s := makeServer(conf, log, appLogic)
		done := make(chan error)
		go startHTTP(s, log, done)
		<-ch
		log.LogInfo("Stopping web server")
		if err := s.Close(); err != nil {
			log.LogInfo("Web server stopping", "error", err)
		}
		if err := s.Shutdown(context.Background()); err != nil {
			log.LogInfo("Web server shutting down", "error", err)
		}
		if err := <-done; err != nil {
			return err
		}
		msg := <-ch
		if !msg.restart {
			log.LogInfo("Web server stopped")
			return nil
		}
		conf = msg.cnf
		appLogic = msg.appLogic
	}
}

func startHTTP(s *http.Server, log logger.T, done chan<- error) {
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.LogError("HTTP server", "error", err)
		if logErr := log.Close(); logErr != nil {
			done <- fmt.Errorf(
				"cannot close HTTP server, error: %s, and cannot close log file: %s",
				err, logErr)
		} else {
			done <- fmt.Errorf("cannot close HTTP server, error: %s", err)
		}
	} else {
		done <- nil
	}
}

func readConfig(confFile string) (config.T, logic.Option, []logic.Option,
	logger.T, error) {
	conf, err := config.Read(confFile)
	if err != nil {
		return config.T{}, logic.Option{}, nil, nil, err
	}

	log,
		err := logger_mux.New(conf.Log)
	if err != nil {
		return config.T{}, logic.Option{}, nil, nil, err
	}

	defaultAppLogic,
		err := getNewAppLogic(log, config.SubT{
		T: config.T{
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
		return config.T{}, logic.Option{}, nil, nil, err
	}

	optionalAppLogic := make([]logic.Option, 0)
	for _, v := range conf.SubConfig {
		opt, err := getNewAppLogic(log, v)
		if err != nil {
			return config.T{}, logic.Option{}, nil, nil, err
		}
		optionalAppLogic = append(optionalAppLogic, opt)
	}
	return conf,
		defaultAppLogic,
		optionalAppLogic,
		log,
		nil
}

func signalsCatcher(confFile string, log logger.T, ch chan<- confChan) {
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
				if err := log.Close(); err != nil {
					log.LogError("log file close", "error", err)
					ch <- confChan{config.T{}, nil, false}
				} else {
					ch <- confChan{conf, logic.New(def, opts), true}
					log = logNew
				}
			}
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			log.LogWarning("Exiting")
			ch <- confChan{}
			if err := log.Close(); err != nil {
				log.LogError("log file close", "error", err)
			}
			ch <- confChan{}
		}
	}
}

func getNewAppLogic(log logger.T, v config.SubT) (logic.Option, error) {
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
	_extractor,
		err := extractor_mux.New(v.Extractor,
		logger_mux.NewLayer(log, newName(texts[0])))
	if err != nil {
		return logic.Option{}, nameErr(texts[0], err)
	}
	_cache,
		err := cache_mux.New(v.Cache,
		logger_mux.NewLayer(log, newName(texts[1])))
	if err != nil {
		return logic.Option{}, nameErr(texts[1], err)
	}
	_streamer,
		err := streamer.New(v.Streamer,
		logger_mux.NewLayer(log, newName(texts[2])), _extractor)
	if err != nil {
		return logic.Option{}, nameErr(texts[2], err)
	}
	return logic.Option{
			Name:               v.Name,
			Sites:              v.Sites,
			X:                  _extractor,
			S:                  _streamer,
			C:                  _cache,
			DefaultVideoHeight: v.DefaultVideoHeight,
			MaxVideoHeight:     v.MaxVideoHeight,
		},
		nil
}

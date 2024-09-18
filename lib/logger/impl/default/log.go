package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	l "lib/logger/config"
)

type loggerT struct {
	lvl *l.LevelT
	lgr *log.Logger
}

func (t *loggerT) print(str string, s string, args []any) {
	var (
		fmtstr string
		arglen = len(args)
		k      int
	)
	add := func(a string) {
		if len(a) > 0 {
			fmtstr = fmt.Sprintf("%s %s", fmtstr, a)
		}
	}
	for k < arglen {
		if k+1 < arglen {
			add(fmt.Sprintf("%s:%+v", args[k], args[k+1]))
			k = k + 2
		} else {
			add(fmt.Sprintf("%+v", args[k]))
			k++
		}
	}
	t.lgr.Println(fmt.Sprintf("%-7s %s.", str, s) + fmtstr)
}

func (t *loggerT) checkAndPrint(lvl l.LevelT, str string, s string, i []any) {
	if *t.lvl <= lvl {
		t.print(str, s, i)
	}
}

func (t *loggerT) LogError(s string, i ...any) {
	t.checkAndPrint(l.Error, "ERROR", s, i)
}
func (t *loggerT) LogWarning(s string, i ...any) {
	t.checkAndPrint(l.Warning, "WARNING", s, i)
}
func (t *loggerT) LogDebug(s string, i ...any) {
	t.checkAndPrint(l.Debug, "DEBUG", s, i)

}
func (t *loggerT) LogInfo(s string, i ...any) {
	t.checkAndPrint(l.Info, "INFO", s, i)
}

func New(conf l.ConfigT) (*loggerT, error) {
	var (
		logger loggerT
		lgr    *log.Logger = log.Default()
	)
	open := func() (*os.File, error) {
		return os.OpenFile(
			// will never close this file :|
			// should trap exit
			*conf.FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	}
	switch *conf.Output {
	case l.Stdout:
		lgr.SetOutput(os.Stdout)
	case l.File:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		lgr.SetOutput(f)
	case l.Both:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		out := io.MultiWriter(os.Stdout, f)
		lgr.SetOutput(out)
	}
	logger.lgr = lgr
	logger.lvl = conf.Level
	return &logger, nil
}

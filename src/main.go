// Package main is entry point
package main

import (
	"flag"
	"os"

	app "ytproxy/app"
	utils "ytproxy/utils"
)

const appVersion = "2.3.1"

type flagsT struct {
	version bool
	config  string
}

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
		utils.WriteStdoutLn(appVersion)
		return
	}
	if err := app.Run(flags.config); err != nil {
		utils.WriteError(err)
		os.Exit(1)
	}
}

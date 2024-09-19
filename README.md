### What is this repository for? ###

This is yt-dlp based video restreamer, part of another project: https://github.com/mesb1/xupnpd_youtube

### Build ###

`cd cmd && go build`

### Options ###

Run with `--help`

### Exit codes ###

  - 0 - OK
  - 1 - config read/parse error
  - 2 - logger create error
  - 3 - extractor create error
  - 4 - streamer create error
  - 5 - web server error
  - 6 - links cache error
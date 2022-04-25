### What is this repository for? ###

This is part of another project: https://github.com/mesb1/xupnpd_youtube

### Build ###

`go build`

### Options ###

Run with `--help`

### Exit codes ###

  - 1 - config read/parse error
  - 2 - logger create error
  - 3 - extractor create error
  - 4 - streamer create error
  - 5 - web server error

### Config explained ###
do not copypaste, comments are not allowed in this json.
use config.default.json instead.

```jsonc
{
    // web server listen port
    "port": 8080,
    // restreamer config
    "streamer": {
        // show errors in headers (insecure)
        "error-headers": false,
        // do not strictly check video headers
        "ignore-missing-headers": false,
        // do not check video server certificate (insecure)
        "ignore-ssl-errors": false,
        // video file that will be shown on errors
        "error-video": "corrupted.mp4",
        // audio file that will be shown on errors
        "error-audio": "corrupted.mp4"
    },
    // media extractor config
    "extractor": {
        // file path
        "path": "yt-dlp",
        // arguments for extractor mp4 url
        // ",," is args separator, not space
        // {{.HEIGHT}} will be replaced with requested height (360/480/720)
        // {{.URL}} will be replace with requested url
        // also you can use {{.FORMAT}} - requested format (now - only mp4 or m4a)
        "mp4": "-f,,(mp4)[height<={{.HEIGHT}}],,-g,,{{.URL}}",
        // same for m4a
        "m4a": "-f,,(m4a),,-g,,{{.URL}}",
        // args for getting user-agent (not used yet)
        "get-user-agent": "--dump-user-agent"
    },
    // logger config
    "log": {
        // log level
        // debug/info/warning/error/nothing
        "level": "info",
        // log destination
        // stdout/file/both
        "output": "stdout",
        // filename if writing to file
        "filename": "log.txt"
    }
}

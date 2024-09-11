### What is this repository for? ###

This is part of another project: https://github.com/mesb1/xupnpd_youtube

### Build ###

`cd src && go build`

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
        // audio file that will be played on errors
        // dwnlded here youtu.be/_b8KPiT1PxI (suggest your options)
        "error-audio": "failed.m4a",
        // how to set streamer's user-agent
        // request - set from user's request (old default)
        // extractor - set from extractor on app start (default)
        // config - set from config
        "set-user-agent": "extractor",
        // custom user agent used if "set-user-agent" set to "config"
        "user-agent": "Mozilla",
        // proxy for restreamer
        // empty - no proxy
        // "env" - read proxy from environment variables (e.g. HTTP_PROXY="http://127.0.0.1:3128")
        // proxy url - e.g. "socks5://127.0.0.1:9999"
        "proxy": "env"
    },
    // media extractor config
    "extractor": {
        // file path
        "path": "yt-dlp",
        // arguments for extractor
        // args separator is ",,", not space
        // {{.HEIGHT}} will be replaced with requested height (360/480/720)
        // {{.URL}} will be replace with requested url
        // also you can use {{.FORMAT}} - requested format (now - only mp4 or m4a)
        "mp4": "-f,,(mp4)[height<={{.HEIGHT}}],,-g,,{{.URL}}",
        // same for m4a
        "m4a": "-f,,(m4a),,-g,,{{.URL}}",
        // args for getting user-agent
        "get-user-agent": "--dump-user-agent",
        // custom options list to extractor, like proxy, etc.
        // same rules as mp4/m4a
        // HEIGHT/URL/.. templates also can be used 
        "custom-options": [
            "--option1,,value1",
            "--option2",
            "value2",
            "--option3",
            "very long value 3"
        ]
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
    },
    // links cache config
    "cache": {
        // links expire time
        // time units are "s", "m", "h", e.g. "1h10m10s", "10h", "1s"
        // "0s" will disable cache
        "expire-time": "3h"
    }   
}

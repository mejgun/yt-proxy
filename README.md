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

### Config explained ###
use config.default.json to start

turn on debug:

```
"debug": false
```
show errors in headers (insecure):

```
"error-headers": false
```
do not strictly check video headers:

```
"ignore-missing-headers": false
```
do not check video server certificate (insecure):

```
"ignore-ssl-errors": false
```
listen port:

```
"port": 8080
```
file that will be shown on errors:

```
"error-video": "corrupted.mp4"
```
Media extractor to use (TODO):

```
"extractor": {}
```
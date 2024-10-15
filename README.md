### About 
Re-streamer application for video/audio streams. It gets link that can't be handled by your device/software and restreams it with regular http.

Main purpose is to allow seamless integration of a different incompatible software and hardware.

It was created because of community request mainly.

### What is this repository for?

This is part of another project: https://github.com/mesb1/xupnpd_youtube

Popular use cases: 
*  hardware players missing support for modern services like youtube, but compatible with IPTV (selenga hd980, wv t625a lan, etc)
*  software content delivery services that lacks modern services support (xupnpd).

### Features

*  Un-securing links, in case your player missing https support, this software removes https and provide necessary content via pure http.
*  Provide content from modern services, by extracting (3rd party project [yt-dlp](https://github.com/yt-dlp/yt-dlp) is in use) direct link to video and streaming it to your software/hardware player.
*  Proxy supported.

### Build

`cd src && go build`

### Quick Start

*  load fresh yt-dlp binary and put it in your system path
*  make sure media files mp4 and m4a is also under the same directory where you run the app
*  fix default config for custom options and my-extractor part (commentout if needed)
*  start the yt-proxy app binary
*  open your favorite player with this link:
  
http://127.0.0.1:8080/play/www.youtube.com/watch?v=9lNZ_Rnr7Jc?/?vh=360&vf=mp4

| URL part  | Description |
| --- | --- |
| `http://127.0.0.1:8080`  | server address where this app is running |
| `/play/`  | required |
| `www.youtube.com/watch?v=9lNZ_Rnr7Jc` | the real video URL, http(s) scheme is optional |
| `?/?` | delimiter, next will be this app options, all are optional |
| `vh=360` | requested video height |
| `&` | options delimiter | 
| `vf=mp4` | requested format, only mp4 and m4a are supported by now |

### Options

Run with `--help`

### Memory usage fine-tuning

You can set standard Go [environment variables](https://pkg.go.dev/runtime#hdr-Environment_Variables), but this may impact the performance.
* [`GOGC`](https://pkg.go.dev/runtime/debug#SetGCPercent) - (default 100) a garbage collector is triggered when memory usage increased by this percentage (roughly speaking)
*  [`GOMEMLIMIT`](https://pkg.go.dev/runtime/debug#SetMemoryLimit) - a soft memory limit (e.g. 500MiB)
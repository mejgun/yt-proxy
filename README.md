### About 
Re-streamer application for video/audio streams. It gets link that can't be handled by your device/software and restreams it with regular http.

Main purpose is to allow seamless integration of a different incompatible software and hardware.

It was created because of community request mainly.

### What is this repository for? ###

This is part of another project: https://github.com/mesb1/xupnpd_youtube

Popular usecases: 
*  hardware players missing support for modern services like youtube, but compatible with iptv (selenga hd980, wv t625a lan, etc)
*  software content delivery services that lacks modern services support (xupnpd).

### Features

*  Un-securing links, in case your player missing https support, this software removes https and provide necessary content via pure http.
*  Provide content from modern services, by extracting (3rd party project yt-dlp is in use) direct link to video and streaming it to your software/hardware player.
*  Proxy supported.

### Build ###

`cd cmd && go build`

### Quick Start ###

*  load fresh yt-dlp binary and put it in your system path
*  make sure media files mp4 and m4a is also there
*  fix default config for custom options and my-extractor part (commentout if needed)
*  start the yt-proxy app binary
*  open your favourite player with this link:
  
http://127.0.0.1:8080/play/www.youtube.com/watch?v=9lNZ_Rnr7Jc?/?vh=360


### Options ###

Run with `--help`

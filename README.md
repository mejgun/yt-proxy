### About 

Re-streamer application for video/audio streams, main purpose is to allow seamless integration of a different incompatible software and hardware.

It was created because of community request mainly.

### What is this repository for? ###

This is part of another project: https://github.com/mesb1/xupnpd_youtube

Popular usecases: hardware players missing support for modern services like youtube, but compatible with iptv (selenga hd980, wv t625a lan, etc), software content delivery services lack modern services support (xupnpd).

### Features

Un-securing links, in case your player missing https support, this software removes https and provide necessary content via pure http.
Provide content from modern services, by extracting (3rd party project yt-dlp is in use) direct link to video and streaming it to your software/hardware player. Proxy supported

### Build ###

`cd cmd && go build`

### Options ###

Run with `--help`
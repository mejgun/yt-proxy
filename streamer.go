package main

import (
	"fmt"
	"io"
	//"bytes"
	//"io/ioutil"
	"log"
	"net/http"
	"os"
)

var c chan RChan

func HelloServer(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req.UserAgent())
	//fmt.Println(req.Cookies())
	fmt.Println(req.Write(os.Stdout))
	url := req.URL.Path[len("/play/"):] + "?"
    url += req.URL.RawQuery
    fmt.Println(url)
	qw := make(chan Response)
	c <- RChan{url: url, c: qw}
	r := <-qw
    fmt.Println(r)
	if r.err == nil {
		// url := "https://r2---sn-pivhx-n8me.googlevideo.com/videoplayback?requiressl=yes&initcwndbps=215000&gcr=ru&mime=video%2Fmp4&key=yt6&mt=1507612964&ratebypass=yes&dur=11495.015&mn=sn-pivhx-n8me&lmt=1507410342958115&ei=dlncWev3E8zqYfm0m6AF&ms=au&ipbits=0&pl=21&mv=m&source=youtube&ip=78.138.154.79&id=o-ADgW5efyBvuR5_jI9bjgph_Nt0ZPYF2fMavvIv_1nYlA&mm=31&expire=1507634646&itag=22&sparams=dur%2Cei%2Cgcr%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Cratebypass%2Crequiressl%2Csource%2Cexpire&signature=1A1E082D5749760CF0F929BD752F10A4FA550022.C1096456EF23883383519265C764A97D15E257E4"
		request, _ := http.NewRequest("GET", r.url, nil)
		r1, ok := req.Header["Range"]
		if ok {
			request.Header.Set("Range", r1[0])
		}
		request.Header.Set("User-Agent", req.UserAgent())
		tr := &http.Transport{}
		client := &http.Client{Transport: tr}
		fmt.Println(request)
		res, _ := client.Do(request)
		defer res.Body.Close()
		h1, ok := res.Header["Content-Length"]
		if ok {
			w.Header().Set("Content-Length", h1[0])
		}
		h2, ok := res.Header["Content-Type"]
		if ok {
			w.Header().Set("Content-Type", h2[0])
		}
		h3, ok := res.Header["Accept-Ranges"]
		if ok {
			w.Header().Set("Accept-Ranges", h3[0])
		}
		if res.StatusCode == 206 {
			w.WriteHeader(http.StatusPartialContent)
		}
		// w.Header().Set("Close", "true")
		//	fmt.Printf("%v\n", res.Header)
		fmt.Printf("%+v\n", res)
		io.Copy(w, res.Body)
		//io.Copy(ioutil.Discard, res.Body)
		/*for _, err := io.CopyN(w, res.Body, 640000); err == nil; {
			fmt.Print(".")
		}*/
		res.Body.Close()
		fmt.Println("---")
		//w.Close()
	}
}

func main() {
	c = make(chan RChan)
    links = make(Links)
	go parseLinks(c)

	var port string
	if len(os.Args) == 2 {
		port = os.Args[1]
	} else {
		port = "8080"
	}
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/play/", HelloServer)
	s := &http.Server{
		Addr: ":" + port,
	}
	s.SetKeepAlivesEnabled(true)
	fmt.Printf("starting at *:%s\n", port)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("HTTP server start failed: ", err)
	}
	//	http.ListenAndServe(":8181", nil)
}

//buf := make([]byte, 6400)
/*for n, err := io.ReadAtLeast(res.Body, buf, 1); err == nil; {
    //robots, _ := ioutil.ReadAll(res.Body)
    //w.Write(robots)
    // fmt.Print(n, " ")
    buf = buf[:n]
    //fmt.Printf("%v %s",n,buf)
    fmt.Printf("%v ", n)
    //w.Write(buf)
    io.Copy(buf, w)

}*/

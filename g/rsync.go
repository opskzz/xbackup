package g

import (
	"io"
	"log"
	"net/http"
	"os"
)

//Download file from redis queue
func Download(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.CopyN(out, resp.Body, SpeedLimit)
	if err != nil {
		return err
	}
	return nil
}

//HTTPRsync fileserver for rsync download
func HTTPRsync(addr, port, dirs string) {
	fs := http.FileServer(http.Dir(dirs))
	http.Handle("/", http.StripPrefix(dirs, fs))
	log.Printf("HTTP Rsync server listen on: %s, port: %s", addr, port)
	http.ListenAndServe(addr+":"+port, nil)
}

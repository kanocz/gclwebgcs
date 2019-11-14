package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
)

// just small sugar for beautifull code :)
func fatalCond(cond bool, msg ...interface{}) {
	if !cond {
		return
	}
	log.Fatalln(msg...)
}

var (
	bucket *storage.BucketHandle
	ctx    = context.Background()
	index  = "index.html"
	n404   = "/404.html"
)

func httpMain(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path[1:]
	if "" == uri {
		uri = index
	}

	obj := bucket.Object(uri)
	objAttrs, err := obj.Attrs(ctx)

	switch err {
	case storage.ErrBucketNotExist:
		fmt.Println("Bucket " + os.Getenv("GCS") + " not found!")
		http.Error(w, "Invalid storage", 500)
		return
	case storage.ErrObjectNotExist:
		if r.URL.Path != n404 {
			http.Redirect(w, r, n404, http.StatusFound)
		} else {
			http.Error(w, "File not found", 404)
		}
		return
	default:
		fmt.Println("Error accesing GCS:", err)
		http.Error(w, "Internal server error", 500)
		return
	case nil:
	}

	obj = obj.ReadCompressed(true)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		fmt.Println("Error accesing GCS:", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", objAttrs.ContentType)
	w.Header().Set("Content-Encoding", objAttrs.ContentEncoding)
	w.Header().Set("Content-Length", strconv.Itoa(int(objAttrs.Size)))
	w.Header().Set("Content-Disposition", objAttrs.ContentDisposition)
	w.Header().Set("Cache-Control", objAttrs.CacheControl)
	w.Header().Set("ETag", objAttrs.Etag)
	w.WriteHeader(200)

	if _, err := io.Copy(w, reader); nil != err {
		fmt.Println("Error on sending file:", err)
	}
}

func main() {

	listen := ":" + os.Getenv("PORT")
	if ":" == listen {
		listen = ":8080"
	}

	fatalCond("" == os.Getenv("GCS"), "No GCS bucket specified")

	client, err := storage.NewClient(ctx)
	fatalCond(nil != err, "error GCS init:", err)

	bucket = client.Bucket(os.Getenv("GCS"))

	http.HandleFunc("/", httpMain)
	http.ListenAndServe(listen, nil)
}

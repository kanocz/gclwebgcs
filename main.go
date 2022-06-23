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

// just small sugar for beautifully code :)
func fatalCond(cond bool, msg ...interface{}) {
	if !cond {
		return
	}
	log.Fatalln(msg...)
}

var (
	bucket      *storage.BucketHandle
	ctx         = context.Background()
	index       = "index.html"
	n404        = "/404.html"
	redirectAll = ""
	allowOrigin = false
)

func httpMain(w http.ResponseWriter, r *http.Request) {
	if "" != redirectAll {
		http.Redirect(w, r, redirectAll+r.URL.Path, http.StatusFound)
		return
	}

	uri := r.URL.Path[1:]
	if "" == uri {
		uri = index
	}

	// in case of folder add index
	if uri[len(uri)-1] == '/' {
		uri += index
	}

	obj := bucket.Object(uri)
	objAttrs, err := obj.Attrs(ctx)

	if storage.ErrObjectNotExist == err {
		tmpURI := uri + "/" + index
		tmpObj := bucket.Object(uri)
		tmpObjAttrs, err := tmpObj.Attrs(ctx)
		if nil == err {
			uri = tmpURI
			obj = tmpObj
			objAttrs = tmpObjAttrs
		}
	}

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
		fmt.Println("Error accessing GCS:", err)
		http.Error(w, "Internal server error", 500)
		return
	case nil:
	}

	obj = obj.ReadCompressed(true)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		fmt.Println("Error accessing GCS:", err)
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
	if allowOrigin {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
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

	redirectAll = os.Getenv("REDIRECT")
	if "" == redirectAll {
		fatalCond("" == os.Getenv("GCS"), "No GCS bucket specified")
		client, err := storage.NewClient(ctx)
		fatalCond(nil != err, "error GCS init:", err)
		bucket = client.Bucket(os.Getenv("GCS"))
	}

	// additional configuration via env variables
	if "true" == os.Getenv("CORS") {
		allowOrigin = true
	}
	if "" != os.Getenv("INDEX") {
		index = os.Getenv("INDEX")
	}
	if "" != os.Getenv("404") {
		n404 = os.Getenv("404")
	}

	http.HandleFunc("/", httpMain)
	http.ListenAndServe(listen, nil)
}

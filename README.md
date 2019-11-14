# gclwebgcs
Simple webserver for Google Cloud Run to server files from Google Cloud Storage


to publish do something like this:
```sh
gcloud builds submit --tag gcr.io/project1/gclwebgcs

gcloud beta run deploy --image gcr.io/project1/gclwebgcs --platform managed --set-env-vars=GCS=GCSbucketName serviceName
```

It's also some ENV values additional to `GSC`:
* `CORS` in case of value `true` will add `Access-Control-Allow-Origin=*` header
* `INDEX` specifies main index of site (/ page), default is `index.html`
* `404` specifies redirect to page in case of unexistent file, default is `/404.html`
* `REDIRECT` is just total unconditional redirect to other domain (!), for example in case of domain change. Value of `REDIRECT` acts as __prefix__ and URI is added. For example if actual domain is example.com and value is "https://example2.com" visitor of `https://example.com/hello/world` will be redirected to `https://example2.com/hello/world` (in case if you don't need it just put something like `https://example2.com/#` to `REDIRECT` and rest part will be almost ignored by browser :) )
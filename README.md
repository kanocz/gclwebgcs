# gclwebgcs
Simple webserver for Google Cloud Run to server files from Google Cloud Storage


to publish do something like this:
```sh
gcloud builds submit --tag gcr.io/project1/gclwebgcs

gcloud beta run deploy --image gcr.io/project1/gclwebgcs --platform managed --set-env-vars=GCS=GCSbucketName serviceName
```

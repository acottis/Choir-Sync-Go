# Choir Sync Go

## Dev Setup

* To authenticate to gcloud while testing locally run `gcloud auth application-default login`
* To run server locally `go run main.go`
* Default port is `8080` unless you set the `PORT` environment variable

## Deploy

1. `gcloud app deploy`

## Cloud Storage Cors

Issue with the media player required us to set a more permissive cors policy. The below is used to update the policy as required.
`gcloud storage buckets update gs://BUCKET_NAME --cors-file=cors.json`

## TODO

* Implement pipeline
* Test deployed

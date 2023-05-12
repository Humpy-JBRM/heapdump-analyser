# Heapdump Analyser Service

A web service to parse and analyse java heap dumps.

## Running on Docker

	- make heapdump docker

	- docker run -p 8000:8000 heapdump:latest

and the heapdump analyser will be running on http://localhost:8000.


## Installing (VPS or server)
First, make sure that you have OpenJDK and MAT installed.

Update the config.yml with the full path to ParseHeapdump.sh in MAT.

Install go 1.20 from https://golang.org.

Compile the binary:

`go build -o heapdump .`

## Running

`./heapdump heapdump`

By default, the service runs on *:8000.  This can be changed in the config.yml.

## Operation

The endpoint of the heapdump analyser is `http://localhost:8000/api/heapdump`

This endpoint takes a `POST` request, with a content type of `multipart/form-data`.

This form should have a single input called `file`, of type *file*.  This file should be a `hprof` binary heapdump of a running JVM.

## Testing

`curl -o /tmp/foo.png -F file=@your-heapdump.hprof http://localhost:8000/api/heapdump` 

You will get a response telling you the job ID for processing this heapdump:

```
{
   "id" : "f8fb8840-1b34-4a64-9e8c-de6e7578366e",
   "queued_millis" : 1683912892201,
   "zipfile" : "/var/folders/_6/qs_199nn7zx9563gmrw440x80000gn/T/1164096443/1411673626-1683897697.hprof"
}
```

when all the processing is done, you can view the heapdump analysis reports at http://localhost:8000/heapdump/{ID}

For OOM problems, the *Leak Suspects* report tends to be the most useful.

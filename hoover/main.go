package main

import (
	"flag"
	"heapdump/fswatcher"
	"log"
	"os"
)

var exitOnOOMDump = flag.Bool("exitOnOOMDump", false, "Whether or not to exit when we detect a JVM OOM heapdump")
var analyseHeapDump = flag.Bool("analyse", true, "Whether or not to automatically analyse the heap and produce a leak suspects report")
var parseHeapdumpCommand = flag.String("parseHeapdumpCmd", "mat/ParseHeapDump.sh", "Path of the ParseHeapdump.sh")

func main() {
	flag.Parse()
	dirToWatch := flag.Args()[0]
	//export AWS_ACCESS_KEY_ID=xxx ; export AWS_SECRET_ACCESS_KEY=yyy ; export AWS_S3_BUCKET=zzz ; ./main -analyse=true -parseHeapdumpCmd=/Applications/mat.app/Contents/Eclipse/ParseHeapDump.sh /var/tmp/heapdumps
	os.Setenv("AWS_ACCESS_KEY_ID", "xxx")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "xxx")
	os.Setenv("AWS_S3_BUCKET", "xxx")
	os.Setenv("PARSE_HEAPDUMP_CMD", *parseHeapdumpCommand)
	if os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") == "" {
		log.Printf("INFO: 'AWS_WEB_IDENTITY_TOKEN_FILE' environment is not set, checking AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY")
		if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
			log.Fatalf("ERROR: The 'AWS_ACCESS_KEY_ID' environment variable must be set")
		}
		if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
			log.Fatalf("ERROR: The 'AWS_SECRET_ACCESS_KEY' environment variable must be set")
		}
	} else {
		log.Printf("INFO: Using credentials AWS_WEB_IDENTITY_TOKEN_FILE=%s", os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"))
	}
	if os.Getenv("AWS_S3_BUCKET") == "" {
		log.Fatalf("ERROR: The 'AWS_S3_BUCKET' environment variable must be set")
	}

	watcher, err := fswatcher.NewHeapdumpWatcher(
		dirToWatch,
		*analyseHeapDump,
		*exitOnOOMDump,
	)
	if err != nil {
		log.Fatal(err)
	}
	watcher.Start()
	select {}
}

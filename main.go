package main

import (
	"flag"
	"io/ioutil"
	"log"
	"regexp"
)

var (
	flagRegex          = `(http[s]?:\/\/.*\.(gif|png|jpg|jpeg))(\W)`
	flagDestinationURL = "https://cdn.kotatsu.club"
	flagBlobStorageURL = "https://kotatsuclubassets.blob.core.windows.net/$web"
	flagDirectory      = "content/post"
	flagDryRun         = false
	flagMaxFileSize    = 8388608
)

func init() {
	flag.StringVar(&flagDirectory, "d", flagDirectory, "directory")
	flag.StringVar(&flagDestinationURL, "url", flagDestinationURL, "endpoint to migrate to")
	flag.StringVar(&flagBlobStorageURL, "b", flagBlobStorageURL, "azure blob storage URL")
	flag.StringVar(&flagRegex, "r", flagRegex, "regex used to find changes")
	flag.BoolVar(&flagDryRun, "dry-run", flagDryRun, "don't change files or perform uploads")
	flag.IntVar(&flagMaxFileSize, "max-bytes", flagMaxFileSize, "max content-length to allow for a upload in bytes")

	flag.Parse()
}

func main() {
	regex := regexp.MustCompile(flagRegex)

	migrator, err := NewMigrator(flagBlobStorageURL, flagMaxFileSize)
	if err != nil {
		log.Fatal(err)
	}

	fileInfos, err := ioutil.ReadDir(flagDirectory)
	if err != nil {
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, fileInfo := range fileInfos {
		file := NewFile(flagDirectory + "/" + fileInfo.Name())

		changes, err := file.FindChanges(regex, flagDestinationURL)
		if err != nil {
			log.Fatal(err)
		}

		numChanges := len(changes)
		if numChanges < 1 {
			continue
		}

		log.Printf("found %d asset(s) to migrate in file: %s\n", numChanges, file.Name)

		for _, change := range changes {
			err := migrator.Migrate(*change, flagDryRun)
			if err != nil {
				log.Fatal(err)
			}
		}

		if !flagDryRun {
			err = file.ApplyChanges(changes)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println("done")
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type Migrator struct {
	maxFileSize  int
	storageURL   url.URL
	containerURL azblob.ContainerURL
}

func NewMigrator(azBlobStorageEndpoint string, maxFileSize int) (*Migrator, error) {
	credential, err := azblob.NewSharedKeyCredential(os.Getenv("AZ_STORAGE_ACCOUNT"), os.Getenv("AZ_STORAGE_ACCESS_KEY"))
	if err != nil {
		return nil, fmt.Errorf("failed to generate azblob shared key credential: %s", err.Error())
	}

	azBlobStorageURL, err := url.Parse(azBlobStorageEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	return &Migrator{
		maxFileSize:  maxFileSize,
		containerURL: azblob.NewContainerURL(*azBlobStorageURL, pipeline),
	}, nil
}

func (m *Migrator) Migrate(change Change, dryRun bool) error {
	if dryRun {
		log.Println("dry-run enabled: skipping migration: " + change.OldURL + " -> " + change.NewURL)
		return nil
	}

	log.Println("processing migration " + change.OldURL + " -> " + change.NewURL)

	resp, err := m.download(change.OldURL)
	if err != nil {
		return fmt.Errorf("failed to download source: %s", err.Error())
	}
	defer resp.Body.Close()

	blobURL := m.containerURL.NewBlockBlobURL(change.NewFilename)

	_, err = azblob.UploadStreamToBlockBlob(context.Background(), resp.Body, blobURL, azblob.UploadStreamToBlockBlobOptions{
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: resp.Header.Get("Content-Type"),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload stream: %s", err.Error())
	}

	return nil
}

func (m *Migrator) download(sourceURL string) (*http.Response, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return nil, err
	}

	contentLengthHeader := resp.Header.Get("Content-Length")

	if contentLengthHeader != "" {
		len, err := strconv.Atoi(contentLengthHeader)
		if err != nil {
			return nil, err
		}

		if len > m.maxFileSize {
			return nil, fmt.Errorf("file size is greater than %d bytes", m.maxFileSize)
		}
	}

	return resp, nil
}

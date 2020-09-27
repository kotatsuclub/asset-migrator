package main

import (
	"strings"

	"github.com/teris-io/shortid"
)

type Change struct {
	OldURL      string
	NewURL      string
	NewFilename string
}

func (c *Change) Apply(content string) string {
	return strings.Replace(content, c.OldURL, c.NewURL, -1)
}

type ChangeRequest struct {
	OldURL     string
	NewBaseURL string
	Extension  string
}

func (c *ChangeRequest) GenerateChange() (*Change, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	newFilename := id + "." + c.Extension
	newURL := c.NewBaseURL + "/" + newFilename

	return &Change{
		OldURL:      c.OldURL,
		NewURL:      newURL,
		NewFilename: newFilename,
	}, nil
}

package main

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
)

type File struct {
	Name    string
	content string
	loaded  bool
}

func NewFile(name string) *File {
	return &File{Name: name}
}

func (f *File) Content() (string, error) {
	if !f.loaded {
		err := f.load()
		if err != nil {
			return "", err
		}

		f.loaded = true
	}

	return f.content, nil
}

func (f *File) FindChanges(regex *regexp.Regexp, newURL string) ([]*Change, error) {
	changes := []*Change{}

	content, err := f.Content()
	if err != nil {
		return nil, err
	}

	for _, match := range regex.FindAllStringSubmatch(content, -1) {
		if len(match) < 3 {
			return nil, errors.New("invalid regex capture groups: expected [0:full match, 1:url, 2:extension]")
		}

		url := match[1]
		ext := match[2]

		if strings.HasPrefix(url, newURL) {
			continue
		}

		changeRequest := ChangeRequest{
			OldURL:     url,
			Extension:  ext,
			NewBaseURL: newURL,
		}

		change, err := changeRequest.GenerateChange()
		if err != nil {
			return nil, err
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (f *File) ApplyChanges(changes []*Change) error {
	content, err := f.Content()
	if err != nil {
		return err
	}

	for _, change := range changes {
		content = change.Apply(content)
	}

	f.content = content

	return f.Save()
}

func (f *File) Save() error {
	return ioutil.WriteFile(f.Name, []byte(f.content), 0)
}

func (f *File) load() error {
	contentBytes, err := ioutil.ReadFile(f.Name)
	if err != nil {
		return err
	}

	f.content = string(contentBytes)
	return nil
}

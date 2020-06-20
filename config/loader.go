package config

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type Loader interface {
	Load() (*Config, error)
}

type fileLoader struct {
	fs       afero.Fs
	filename string
}

type tmplVars struct {
	Env map[string]string
}

func NewFileLoader(fs afero.Fs, filename string) Loader {
	return &fileLoader{
		fs:       fs,
		filename: filename,
	}
}

func (f *fileLoader) Load() (*Config, error) {
	content, err := afero.ReadFile(f.fs, f.filename)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, errors.New("file is empty")
	}

	tmpl, err := template.New("config").Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, tmplVars{
		Env: getEnv(),
	})
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buf.Bytes(), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv() map[string]string {
	env := map[string]string{}

	for _, keyVal := range os.Environ() {
		pair := strings.Split(keyVal, "=")
		env[pair[0]] = pair[1]
	}

	return env
}

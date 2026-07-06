package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"sort"
	"ssh2/models"
	"ssh2/parser"
)

var applyCommand = &cli.Command{
	Name:    "apply",
	Aliases: []string{"create"},
	Usage:   "apply resource definition",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) (err error) {
		return newApplyResolver(ctx.String("file")).applyFile(ctx.String("file"))
	},
}

type applyResolver struct {
	dir       string
	parser    parser.YamlParser
	resolving map[string]bool
}

func newApplyResolver(file string) *applyResolver {
	resolver := &applyResolver{
		dir:       filepath.Dir(file),
		resolving: map[string]bool{},
	}
	resolver.parser = parser.YamlParser{ResolveRef: resolver.resolveRef}
	return resolver
}

func (r *applyResolver) applyFile(file string) error {
	records, err := readDocumentRecords(file)
	if err != nil {
		return err
	}
	for _, record := range records {
		if _, err := r.parser.ParseRecord(record); err != nil {
			return err
		}
	}
	return nil
}

func (r *applyResolver) resolveRef(kind, field string, value interface{}) (models.Model, error) {
	key := fmt.Sprintf("%s:%s:%v", kind, field, value)
	if r.resolving[key] {
		return nil, fmt.Errorf("circular ref %s", key)
	}
	r.resolving[key] = true
	defer delete(r.resolving, key)

	records, err := r.findMatchingRecords(kind, field, value)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		instance, err := r.parser.ParseRecord(record)
		if err != nil {
			return nil, err
		}
		return *instance, nil
	}
	return nil, fmt.Errorf("field %s=%v not exists", field, value)
}

func (r *applyResolver) findMatchingRecords(kind, field string, value interface{}) ([]parser.DocumentRecord, error) {
	files, err := yamlFiles(r.dir)
	if err != nil {
		return nil, err
	}
	var records []parser.DocumentRecord
	for _, file := range files {
		fileRecords, err := readDocumentRecords(file)
		if err != nil {
			return nil, err
		}
		for _, record := range fileRecords {
			if record.Kind != kind || record.Spec == nil {
				continue
			}
			if fmt.Sprint(record.Spec[field]) == fmt.Sprint(value) {
				records = append(records, record)
			}
		}
	}
	return records, nil
}

func readDocumentRecords(file string) ([]parser.DocumentRecord, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var records []parser.DocumentRecord
	for {
		var record parser.DocumentRecord
		err := decoder.Decode(&record)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if record.Kind == "" && record.Spec == nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

func yamlFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

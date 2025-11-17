package util

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"

	goyaml "github.com/goccy/go-yaml"
)

func YamlRelativePath(filePath string) string {
	pwd, err := os.Getwd()
	if err != nil {
		zap.S().Fatal(err)
	}

	rel, err := filepath.Rel(pwd, filePath)
	if err != nil {
		zap.S().Fatal(err)
	}

	return rel
}

func YamlDecodeWithComments(filePath string, raw []byte, out any) (goyaml.CommentMap, error) {
	if err := YamlValidate(filePath, raw, out); err != nil {
		return nil, err
	}

	cm := goyaml.CommentMap{}
	dec := goyaml.NewDecoder(bytes.NewReader(raw), goyaml.CommentToMap(cm))
	if err := dec.Decode(out); err != nil {
		return nil, fmt.Errorf("file %s decode failed: %w", YamlRelativePath(filePath), err)
	}

	return cm, nil
}

func YamlEncodeWithComments(filePath string, out any, cm goyaml.CommentMap) ([]byte, error) {
	data, err := goyaml.MarshalWithOptions(out, goyaml.WithComment(cm))
	if err != nil {
		return nil, fmt.Errorf("file %s encode failed: %w", YamlRelativePath(filePath), err)
	}

	return data, nil
}

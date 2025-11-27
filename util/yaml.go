package util

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"

	goyaml "github.com/goccy/go-yaml"
)

func YAMLRelativePath(filePath string) string {
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

func YAMLDecodeWithComments(filePath string, raw []byte, out any) (goyaml.CommentMap, error) {
	if err := YAMLValidate(filePath, raw, out); err != nil {
		return nil, err
	}

	cm := goyaml.CommentMap{}
	dec := goyaml.NewDecoder(bytes.NewReader(raw), goyaml.CommentToMap(cm))
	if err := dec.Decode(out); err != nil {
		return nil, fmt.Errorf("file %s decoding failed: %w", YAMLRelativePath(filePath), err)
	}

	return cm, nil
}

func YAMLEncodeWithComments(filePath string, out any, cm goyaml.CommentMap) ([]byte, error) {
	data, err := goyaml.MarshalWithOptions(out, goyaml.WithComment(cm))
	if err != nil {
		return nil, fmt.Errorf("file %s encoding failed: %w", YAMLRelativePath(filePath), err)
	}

	return data, nil
}

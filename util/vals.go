package util

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/helmfile/vals"
)

const (
	valsCacheSize = 256
)

var instance *vals.Runtime
var once sync.Once

func ValsInstance() (*vals.Runtime, error) {
	var err error
	once.Do(func() {
		instance, err = vals.New(vals.Options{CacheSize: valsCacheSize, LogOutput: io.Discard})
	})

	return instance, err
}

func ValsFetchSecretValue(path string) (string, error) {
	valsMap := make(map[string]any)
	valsMap["key"] = path
	resultMap, err := ValsExpandSecretRefs(valsMap)
	if err != nil {
		return "", err
	}

	rendered, ok := resultMap["key"]
	if !ok {
		return "", fmt.Errorf("unexpected error occurred, %v doesn't have 'key' key", resultMap)
	}

	result, ok := rendered.(string)
	if !ok {
		return "", fmt.Errorf("expected %v to be string", rendered)
	}

	return result, nil
}

func ValsExpandSecretRefs(values map[string]any) (map[string]any, error) {
	// Specific env vars to disable AWS SDK v2 debug logging
	awsEnvs := map[string]string{
		"AWS_SDK_LOAD_CONFIG":  "1",
		"AWS_SDK_GO_LOG_LEVEL": "off",
	}

	if err := SetOSEnvs(false, awsEnvs); err != nil {
		return nil, err
	}

	runtime, err := ValsInstance()
	if err != nil {
		return nil, err
	}

	return runtime.Eval(values)
}

func ValsContainsRefPart(u, part string) bool {
	if !strings.HasPrefix(u, "ref+") {
		return false
	}

	i := strings.Index(u, "://")
	if i == -1 {
		return false
	}

	rest := u[i+3:] // after "://"
	if q := strings.IndexByte(rest, '?'); q != -1 {
		rest = rest[:q] // strip query parameters
	}

	return strings.Contains(rest, part)
}

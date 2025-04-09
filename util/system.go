package util

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type SpecCMD struct {
	Args          []string
	Envs          []string
	Command       string
	Dir           string
	Ctx           *cli.Context
	StdoutBuf     bytes.Buffer
	StderrBuf     bytes.Buffer
	CommandStr    string
	DisableStdOut bool
	Debug         bool
	SensKeyWords  []string
}

func (s *SpecCMD) AddOSEnv() error {
	path, exists := os.LookupEnv("PATH")
	if exists {
		if err := os.Setenv("PATH", GetHomePath(ToolsLocalDir, ToolsBinDir)+":"+path); err != nil {
			return err
		}
	}

	s.Envs = append(os.Environ(), s.Envs...)

	return nil
}

func (s *SpecCMD) sensitive(data []byte) ([]byte, error) {
	for _, word := range s.SensKeyWords {
		regex, err := regexp.Compile(word)
		if err != nil {
			return nil, err
		}

		data = regex.ReplaceAllLiteral(data, []byte("[rmk_sensitive]"))
	}

	return data, nil
}

func (s *SpecCMD) copyAndCapture(r io.Reader, w ...io.Writer) error {
	var errSens error
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			data := buf[:n]
			data, errSens = s.sensitive(data)
			if errSens != nil {
				return errSens
			}

			for _, val := range w {
				if _, err := val.Write(data); err != nil {
					return err
				}
			}
		}

		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}

			return err
		}
	}
}

func (s *SpecCMD) disableStdOut(w ...io.Writer) []io.Writer {
	if s.DisableStdOut {
		return w[:1]
	}

	return w
}

func (s *SpecCMD) ExecCMD() error {
	var (
		wg                 sync.WaitGroup
		stdoutIn, stderrIn io.ReadCloser
		err                error
	)

	if s.Ctx.Command != nil {
		s.Envs = append(s.Envs, "RMK_COMMAND_CATEGORY="+s.Ctx.Command.Category)
	}

	cmd := exec.CommandContext(s.Ctx.Context, s.Command, s.Args...)
	cmd.Dir = s.Dir
	cmd.Env = s.Envs
	if stdoutIn, err = cmd.StdoutPipe(); err != nil {
		return err
	}

	if stderrIn, err = cmd.StderrPipe(); err != nil {
		return err
	}

	cmd.Stdin = os.Stdout

	s.CommandStr = cmd.String()

	err = cmd.Start()
	if err != nil {
		return err
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	wg.Add(1)
	go func() {
		if err = s.copyAndCapture(stdoutIn, s.disableStdOut(&s.StdoutBuf, os.Stdout)...); err != nil {
			zap.S().Fatal(err)
		}

		wg.Done()
	}()

	if err = s.copyAndCapture(stderrIn, s.disableStdOut(&s.StderrBuf, os.Stderr)...); err != nil {
		return err
	}

	wg.Wait()

	return cmd.Wait()
}

func GetHomePath(path ...string) string {
	var absPath []string

	home, err := os.UserHomeDir()
	if err != nil {
		zap.S().Fatal(err)
	}

	return filepath.Join(append(append(absPath, home), path...)...)
}

func GetPwdPath(path ...string) string {
	var absPath []string
	pwd, err := os.Getwd()
	if err != nil {
		zap.S().Fatal(err)
	}

	return filepath.Join(append(append(absPath, pwd), path...)...)
}

func IsExists(path string, file bool) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	if file {
		return !info.IsDir()
	} else {
		return info.IsDir()
	}
}

func FindDir(path, pattern string, byPrefix, bySuffix bool) string {
	fileInfo, err := os.ReadDir(path)
	if err != nil {
		zap.S().Fatal(err)
	}

	for _, dir := range fileInfo {
		if dir.IsDir() {
			switch {
			case byPrefix && strings.HasPrefix(dir.Name(), pattern):
				return dir.Name()
			case bySuffix && strings.HasSuffix(dir.Name(), pattern):
				return dir.Name()
			case dir.Name() == pattern:
				return dir.Name()
			}
		}
	}

	return ""
}

func ListDir(path string, abs bool) (dirs []string, files []string, err error) {
	var pathName string

	infoFiles, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}

	for _, info := range infoFiles {
		if abs {
			pathName = filepath.Join(path, info.Name())
		} else {
			pathName = info.Name()
		}

		if !info.IsDir() {
			files = append(files, pathName)
		}

		if info.IsDir() {
			dirs = append(dirs, pathName)
		}
	}

	return
}

func WalkMatch(rootPath, pattern string) ([]string, error) {
	var match []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			match = append(match, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return match, nil
}

func WalkInDir(rootPath, dir, name string) ([]string, error) {
	var match []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == dir && IsExists(filepath.Join(path, name), true) {
			match = append(match, filepath.Join(path, name))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return match, nil
}

func CopyDir(src, dst string) error {
	rootDir := filepath.Base(src)
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		relPath := strings.Replace(path, src, "", 1)
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rootDir, relPath), 0755)
		} else {
			data, err := os.ReadFile(filepath.Join(src, relPath))
			if err != nil {
				return err
			}

			return os.WriteFile(filepath.Join(dst, rootDir, relPath), data, 0755)
		}
	})
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0755)
}

func MergeAgeKeys(dir string) error {
	var keys []byte

	match, err := WalkMatch(dir, "*"+SopsAgeKeyExt)
	if err != nil {
		return err
	}

	for _, path := range match {
		if filepath.Base(path) != SopsAgeKeyFile {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			keys = append(keys, data...)
		}
	}

	return os.WriteFile(filepath.Join(dir, SopsAgeKeyFile), keys, 0644)
}

func ReadStdin(text string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %s: ", text)
	value, _, err := reader.ReadLine()
	if err != nil {
		zap.S().Fatal(err)
	}

	return string(value)
}

// UnTar takes a destination path and a reader; a tar reader loops over the tar file
// creating the file structure at 'dst' along the way, and writing any files
func UnTar(dst, excludeRegexp string, r io.Reader) error {
	var reg *regexp.Regexp

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	defer gzr.Close()

	if len(excludeRegexp) > 0 {
		reg = regexp.MustCompile(excludeRegexp)
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		matchRegexp := false
		if len(excludeRegexp) > 0 {
			matchRegexp = reg.MatchString(header.Name)
		}

		switch {
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		// if the header is matching with exclude regexp, skip file or dir for unpackage
		case matchRegexp:
			continue
		}
		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// check the file type
		switch header.Typeflag {
		// if it's a dir, and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			if IsExists(target, true) {
				if err := os.Truncate(target, 0); err != nil {
					return err
				}
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			// manually close here after each file operation; deferring would cause each file close
			// to wait until all operations have completed.
			_ = f.Close()
		}
	}
}

func CreateTempYAMLFile(dirPath, fileName string, content []byte) (string, error) {
	file, err := os.CreateTemp(dirPath, fileName+".*.yaml")
	if err != nil {
		return "", err
	}

	if _, err := file.Write(content); err != nil {
		return "", err
	}

	defer file.Close()

	return file.Name(), nil
}

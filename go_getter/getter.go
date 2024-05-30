package go_getter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/hashicorp/go-getter"
	"go.uber.org/zap"
)

type ProgressBar struct {
	lock sync.Mutex
	pool *pb.Pool
	pbs  int
}

var defaultProgressBar getter.ProgressTracker = &ProgressBar{}

func ProgressBarConfig(bar *pb.ProgressBar) {
	bar.SetUnits(pb.U_BYTES)
}

func (cpb *ProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()

	emptyReturn := &readCloser{Reader: stream, close: func() error { return nil }}

	if totalSize <= 0 {
		return emptyReturn
	}

	newPb := pb.New64(totalSize)
	newPb.Set64(currentSize)
	ProgressBarConfig(newPb)
	if cpb.pool == nil {
		cpb.pool = pb.NewPool()
		err := cpb.pool.Start()
		if err != nil {
			return emptyReturn
		}
	}

	cpb.pool.Add(newPb)
	reader := newPb.NewProxyReader(stream)
	cpb.pbs++

	return &readCloser{
		Reader: reader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()
			newPb.Finish()
			cpb.pbs--
			if cpb.pbs == 0 {
				err := cpb.pool.Stop()
				if err != nil {
					return err
				}
				cpb.pool = nil
			}
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() error {
	return c.close()
}

func DownloadArtifact(src, dst, name string, header *http.Header, silent, progress bool, ctxP context.Context) error {
	ctx, cancel := context.WithCancel(ctxP)

	client := &getter.Client{
		Ctx: ctx,
		//define the destination to where the directory will be stored. This will create the directory if it doesnt exist
		Dst: dst,
		Dir: true,
		//the repository with a subdirectory I would like to clone only
		Src:  src,
		Mode: getter.ClientModeAny,
		//define the type of detectors go getter should use, in this case only github is needed
		//Detectors: []getter.Detector{
		//	&getter.GitHubDetector{},
		//},
		//provide the getter needed to download the files
		Getters: map[string]getter.Getter{
			"file":  new(getter.FileGetter),
			"git":   new(getter.GitGetter),
			"gcs":   new(getter.GCSGetter),
			"hg":    new(getter.HgGetter),
			"s3":    new(getter.S3Getter),
			"http":  &getter.HttpGetter{Header: *header},
			"https": &getter.HttpGetter{Header: *header},
		},
	}

	if progress && !silent {
		client.ProgressListener = defaultProgressBar
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		return fmt.Errorf("signal: %v", sig)
	case <-ctx.Done():
		wg.Wait()
		if !silent {
			zap.S().Infof("downloaded: %s", name)
		}
	case err := <-errChan:
		wg.Wait()
		return fmt.Errorf("error downloading: %s", err)
	}

	return nil
}

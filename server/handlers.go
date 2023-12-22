package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/sftp"
)

// custom handlers so we can back commands off to remote http server rather than the local filesystem

// empty file writer and file cmd handlers as we arent supporting any write and file command operations
// but they need defintions as part of sftp.Handlers
type customFileWriter struct{}

func (c customFileWriter) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	return nil, nil
}

type customFileCmder struct{}

func (c customFileCmder) Filecmd(r *sftp.Request) error {
	return nil
}

// fake a file as an sftp client needs to get the file stats before it will issue a get
// this works fine in a normal setting when sftp is serving from the local filesystem, but as are we
// injecting a custom file read handler which is reading from a remote http server we need to to trick the
// client into issuing the get command by feeding it fake file stats

type fakeFile struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (f *fakeFile) Name() string       { return f.name }
func (f *fakeFile) Size() int64        { return f.size }
func (f *fakeFile) Mode() os.FileMode  { return 0666 }
func (f *fakeFile) ModTime() time.Time { return f.modTime }
func (f *fakeFile) IsDir() bool        { return f.isDir }
func (f *fakeFile) Sys() interface{}   { return nil }

type listerat []os.FileInfo

type customFileLister struct{}

func (l listerat) ListAt(p []os.FileInfo, off int64) (int, error) {
	copied := copy(p, l[off:])
	return copied, nil
}

func (c customFileLister) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	log.Printf("DEBUG: method called: %s", r.Method)

	switch r.Method {
	case "List":
		var files listerat = []os.FileInfo{&fakeFile{name: "fakefile.txt"}}
		return files, nil
	case "Stat":
		var file listerat = []os.FileInfo{&fakeFile{name: "fakefile.txt"}}
		return file, nil
	}

	return nil, errors.New("unsupported operation")
}

type customFileReader struct {
	remoteUrl string
}

func (c customFileReader) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	log.Print("DEBUG: reader called")

	resp, err := http.Get("http://grafana.networks-util.ask4.net:8080/" + r.Filepath)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {

		return nil, err
	}

	return bytes.NewReader(data), nil
}

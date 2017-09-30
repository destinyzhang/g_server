package config

import (
	"io"
	"net/http"
	"os"
)

type IFileStream interface {
	Load() (error, io.ReadCloser)
}

type IFileParse interface {
	Parse(rd io.Reader) error
}

type FileReader struct {
	fileparse IFileParse
	stream    IFileStream
	err       error
}

func (freader *FileReader) Error() error {
	return freader.err
}

func (freader *FileReader) GetIFileParse() IFileParse {
	return freader.fileparse
}

func (freader *FileReader) Attach(fileparse IFileParse, stream IFileStream) {
	freader.fileparse = fileparse
	freader.stream = stream
}

func (freader *FileReader) AttachIFileParse(fileparse IFileParse) {
	freader.fileparse = fileparse
}

func (freader *FileReader) AttachIFileStream(stream IFileStream) {
	freader.stream = stream
}

func (freader *FileReader) LoadFile() error {
	if freader.stream == nil || freader.fileparse == nil {
		panic("stream or fileparse nil!!!!!")
	}
	var readercloser io.ReadCloser
	freader.err, readercloser = freader.stream.Load()
	if freader.err != nil {
		return freader.err
	}
	defer readercloser.Close()
	freader.err = freader.fileparse.Parse(readercloser)
	return freader.err
}

type LocalFileStream struct {
	filename string
}

func (stream *LocalFileStream) Load() (error, io.ReadCloser) {
	file, err := os.Open(stream.filename)
	if err != nil {
		return err, nil
	}
	return nil, file
}

type HttpFileStream struct {
	filename string
}

func (stream *HttpFileStream) Load() (error, io.ReadCloser) {
	resp, err := http.Get(stream.filename)
	if err != nil {
		return err, nil
	}
	return nil, resp.Body
}

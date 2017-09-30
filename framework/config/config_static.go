package config

func NewFileReader(fileparse IFileParse, stream IFileStream) *FileReader {
	return &FileReader{fileparse: fileparse, stream: stream}
}

func NewLocalFileStream(fliename string) *LocalFileStream {
	return &LocalFileStream{filename: fliename}
}

func NewHttpFileStream(fliename string) *HttpFileStream {
	return &HttpFileStream{filename: fliename}
}

func NewFileKvParse() *FileKvParse {
	return &FileKvParse{}
}

func NewFileTableParse() *FileTableParse {
	return &FileTableParse{}
}

package importer

import (
	"os"

	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/errs"
	v0 "xn--gckvb8fzb.com/zeit/helpers/importer/engines/v0"
	v1 "xn--gckvb8fzb.com/zeit/helpers/importer/engines/v1"
)

type ImportFileType string

const (
	TypeZeitV0 ImportFileType = "v0"
	TypeZeitV1 ImportFileType = "v1"
)

type ImportEngine interface {
	Import(
		func(database.Model, error, ...any) error,
		...any,
	) error
}

type Importer struct {
	FileType ImportFileType
	File     string
	fd       *os.File
	Engine   ImportEngine
}

func New(ftype ImportFileType, file string) (*Importer, error) {
	var err error

	im := new(Importer)
	im.FileType = ftype
	im.File = file

	if im.fd, err = os.Open(im.File); err != nil {
		return nil, err
	}

	switch im.FileType {
	case TypeZeitV0:
		im.Engine, err = v0.New(im.fd)
	case TypeZeitV1:
		im.Engine, err = v1.New(im.fd)
	default:
		err = errs.ErrUnknownImportFormat
	}

	if err != nil {
		im.fd.Close()
		return nil, err
	}

	return im, nil
}

func (im *Importer) Import(
	cb func(database.Model, error, ...any) error,
	v ...any,
) error {
	return im.Engine.Import(cb, v...)
}

func (im *Importer) End() {
	im.fd.Close()
}

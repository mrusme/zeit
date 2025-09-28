package importer

import (
	"os"

	"github.com/mrusme/zeit/database"
	v0 "github.com/mrusme/zeit/helpers/importer/engines/v0"
	v1 "github.com/mrusme/zeit/helpers/importer/engines/v1"
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
		if im.Engine, err = v0.New(im.fd); err != nil {
			return nil, err
		}
	case TypeZeitV1:
		if im.Engine, err = v1.New(im.fd); err != nil {
			return nil, err
		}
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

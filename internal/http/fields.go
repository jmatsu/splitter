package http

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type Kind = int8

const (
	ZeroField Kind = iota
	File
	NonFile
)

type ValueField struct {
	FieldName string
	Value     string
	Kind      Kind
}

func FileField(name string, path string) ValueField {
	return ValueField{
		FieldName: name,
		Value:     path,
		Kind:      File,
	}
}

func StringField(name string, value string) ValueField {
	return ValueField{
		FieldName: name,
		Value:     value,
		Kind:      NonFile,
	}
}

func BooleanField(name string, value bool) ValueField {
	return ValueField{
		FieldName: name,
		Value:     fmt.Sprintf("%t", value),
		Kind:      NonFile,
	}
}

func (f *ValueField) Open() (string, io.Reader, error) {
	switch f.Kind {
	case File:
		if ref, err := os.Open(f.Value); err != nil {
			return f.FieldName, nil, err
		} else {
			return f.FieldName, ref, nil
		}
	case NonFile:
		return f.FieldName, strings.NewReader(f.Value), nil
	default:
		panic(fmt.Errorf("unsupported field kind: %v", f.Kind))
	}
}

type Form struct {
	Fields []ValueField
}

func (f *Form) Serialize() (string, *bytes.Buffer, error) {
	var buffer bytes.Buffer

	w := multipart.NewWriter(&buffer)
	//goland:noinspection GoUnhandledErrorResult
	defer w.Close()

	for _, field := range f.Fields {
		err := func() error {
			name, reader, err := field.Open()

			if err != nil {
				return err
			}

			if closable, ok := reader.(io.Closer); ok {
				//goland:noinspection GoUnhandledErrorResult
				defer closable.Close()
			}

			switch field.Kind {
			case File:
				if fw, err := w.CreateFormFile(name, filepath.Base(field.Value)); err != nil {
					return err
				} else if _, err = io.Copy(fw, reader); err != nil {
					return err
				}
			case NonFile:
				if fw, err := w.CreateFormField(name); err != nil {
					return err
				} else if _, err = io.Copy(fw, reader); err != nil {
					return err
				}
			default:
				panic(fmt.Errorf("dead case"))
			}

			return nil
		}()

		if err != nil {
			return "", nil, err
		}
	}

	return w.FormDataContentType(), &buffer, nil
}

package net

import (
	"bytes"
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
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

// ValueField represents a single form field.
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
		logger.Logger.Debug().Msgf("File: %s = %s", f.FieldName, f.Value)

		if ref, err := os.Open(f.Value); err != nil {
			return f.FieldName, nil, err
		} else {
			return f.FieldName, ref, nil
		}
	case NonFile:
		logger.Logger.Debug().Msgf("NonFile: %s = %s", f.FieldName, f.Value)

		return f.FieldName, strings.NewReader(f.Value), nil
	default:
		panic(fmt.Sprintf("unsupported field kind: %v", f.Kind))
	}
}

// Form represents a set of form fields.
type Form struct {
	Fields []ValueField
}

func (f *Form) Empty() bool {
	return len(f.Fields) == 0
}

func (f *Form) Set(field ValueField) {
	if f.Fields == nil {
		f.Fields = []ValueField{field}
	} else {
		f.Fields = append(f.Fields, field)
	}
}

func (f *Form) Serialize() (string, *bytes.Buffer, error) {
	var buffer bytes.Buffer

	w := multipart.NewWriter(&buffer)
	//goland:noinspection GoUnhandledErrorResult
	defer w.Close()

	for _, field := range f.Fields {
		field := field

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
				logger.Logger.Debug().Msgf("serialize %s as file in from", name)
				if fw, err := w.CreateFormFile(name, filepath.Base(field.Value)); err != nil {
					return err
				} else if _, err = io.Copy(fw, reader); err != nil {
					return err
				}
			case NonFile:
				logger.Logger.Debug().Msgf("serialize %s as string in from", name)
				if fw, err := w.CreateFormField(name); err != nil {
					return err
				} else if _, err = io.Copy(fw, reader); err != nil {
					return err
				}
			default:
				panic("dead case")
			}

			return nil
		}()

		if err != nil {
			return "", nil, err
		}
	}

	return w.FormDataContentType(), &buffer, nil
}

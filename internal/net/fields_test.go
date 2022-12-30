package net

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func Test_ValueField_Open(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(os.TempDir(), "splitter")

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	if err != nil {
		panic(err)
	}

	cases := map[string]struct {
		field ValueField
	}{
		"string": {
			field: StringField("field1", "value1"),
		},
		"bool": {
			field: BooleanField("field1", true),
		},
		"file": {
			field: FileField("field1", filepath.Join(tempDir, "file1.txt")),
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if c.field.Kind == File {
				if f, err := os.Create(c.field.Value); err != nil {
					t.Errorf("failed to create the file of %s: %v", name, err)
				} else if _, err := f.WriteString(c.field.Value); err != nil {
					t.Errorf("failed to write the content of %s: %v", name, err)
				}
			}

			if name, reader, err := c.field.Open(); err != nil {
				t.Errorf("failed to open the field of %s: %v", name, err)
			} else {
				if x, ok := reader.(io.Closer); ok {
					//goland:noinspection GoUnhandledErrorResult
					defer x.Close()
				}

				if c.field.FieldName != name {
					t.Errorf("field name is expected to be %s but not: %s", c.field.FieldName, name)
				} else if bytes, err := io.ReadAll(reader); err != nil {
					t.Errorf("failed to read the field of %s: %v", name, err)
				} else if string(bytes) != c.field.Value {
					t.Errorf("value is expected to be %s but not: %s", c.field.Value, name)
				}
			}
		})
	}
}

func Test_Form_Serialize(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(os.TempDir(), "splitter")

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	if err != nil {
		panic(err)
	}

	cases := map[string]struct {
		field ValueField
	}{
		"string": {
			field: StringField("field1", "value1"),
		},
		"bool": {
			field: BooleanField("field1", true),
		},
		"file": {
			field: FileField("field1", filepath.Join(tempDir, "file1.txt")),
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if c.field.Kind == File {
				if f, err := os.Create(c.field.Value); err != nil {
					t.Errorf("failed to create the file of %s: %v", name, err)
				} else if _, err := f.WriteString(c.field.Value); err != nil {
					t.Errorf("failed to write the content of %s: %v", name, err)
				}
			}

			if name, reader, err := c.field.Open(); err != nil {
				t.Errorf("failed to open the field of %s: %v", name, err)
			} else {
				if x, ok := reader.(io.Closer); ok {
					//goland:noinspection GoUnhandledErrorResult
					defer x.Close()
				}

				if c.field.FieldName != name {
					t.Errorf("field name is expected to be %s but not: %s", c.field.FieldName, name)
				} else if bytes, err := io.ReadAll(reader); err != nil {
					t.Errorf("failed to read the field of %s: %v", name, err)
				} else if string(bytes) != c.field.Value {
					t.Errorf("value is expected to be %s but not: %s", c.field.Value, name)
				}
			}
		})
	}
}

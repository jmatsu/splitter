package net

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
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
					defer func() {
						_ = x.Close()
					}()
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
					defer func() {
						_ = x.Close()
					}()
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

func Test_Form_Set(t *testing.T) {
	t.Parallel()

	form := Form{}

	if !form.Empty() {
		t.Errorf("a zero form is expected to be empty but not")
	}

	form.Set(StringField("param1", "value1"))

	if form.Empty() {
		t.Errorf("a form having a field is expected not to be empty but is")
	}

	form.Set(FileField("file1", "path/to/file"))
	form.Set(BooleanField("flag1", true))

	expected := []ValueField{
		StringField("param1", "value1"),
		FileField("file1", "path/to/file"),
		BooleanField("flag1", true),
	}

	if !reflect.DeepEqual(form.Fields, expected) {
		t.Errorf("fields are expected to be %v but %v", expected, form.Fields)
	}
}

func Test_ValueField(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		field    ValueField
		expected ValueField
	}{
		"file": {
			field:    FileField("file1", "path/to/file"),
			expected: ValueField{FieldName: "file1", Value: "path/to/file", Kind: File},
		},
		"string": {
			field:    StringField("param1", "value1"),
			expected: ValueField{FieldName: "param1", Value: "value1", Kind: NonFile},
		},
		"boolean true": {
			field:    BooleanField("flag1", true),
			expected: ValueField{FieldName: "flag1", Value: "true", Kind: NonFile},
		},
		"boolean false": {
			field:    BooleanField("flag1", false),
			expected: ValueField{FieldName: "flag1", Value: "false", Kind: NonFile},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !reflect.DeepEqual(c.field, c.expected) {
				t.Errorf("%s is expected to be %v but %v", name, c.expected, c.field)
			}
		})
	}
}

func Test_Form_Serialize_missingFile(t *testing.T) {
	t.Parallel()

	form := Form{}
	form.Set(FileField("file1", filepath.Join(t.TempDir(), "not-found")))

	if _, _, err := form.Serialize(); err == nil {
		t.Errorf("a missing file is expected to be rejected but not")
	}
}

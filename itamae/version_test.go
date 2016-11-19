package itamaelocal

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestProvisioner_Version(t *testing.T) {
	var err error
	var p Provisioner

	buffer := &bytes.Buffer{}

	log.SetOutput(buffer)
	defer func() {
		buffer.Reset()
		buffer = nil
		log.SetOutput(ioutil.Discard)
	}()

	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := "[INFO] Provisioner Itamae v0.1.0"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect version, given: \"%v\", want \"%v\"",
			buffer.String(), expected)
	}

	buffer.Reset()

	Revision = "some-git-revision-12345"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected = "[INFO] Provisioner Itamae v0.1.0 (some-git-revision-12345)"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect version, given: \"%v\", want \"%v\"",
			buffer.String(), expected)
	}
}

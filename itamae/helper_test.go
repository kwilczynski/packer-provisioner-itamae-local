package itamaelocal

import (
	"io"
	"io/ioutil"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func testUI(writer io.Writer) *packer.MachineReadableUi {
	if writer == nil {
		writer = ioutil.Discard
	}
	return &packer.MachineReadableUi{
		Writer: writer,
	}
}

func testCommunicator() *packer.MockCommunicator {
	return &packer.MockCommunicator{}
}

package main

import (
	"reflect"
	"testing"

	"github.com/kwilczynski/packer-provisioner-itamae-local/itamae"
)

func TestMain(t *testing.T) {
	p := itamaelocal.Provisioner{}

	func(v interface{}) {
		if _, ok := v.(itamaelocal.Provisioner); !ok {
			t.Fatalf("not a itamae.Provisioner type: %s", reflect.TypeOf(v).String())
		}
	}(p)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("should panic")
			return
		}
	}()

	main()
}

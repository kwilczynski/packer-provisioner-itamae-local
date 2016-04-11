/*
 * main_test.go
 *
 * Copyright 2016 Krzysztof Wilczynski
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

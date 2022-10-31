/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package registry

import "testing"

func TestMergeMaps(t *testing.T) {
	a := map[string]interface{}{
		"simple":     0, // to be changed to 1
		"to_delete":  0,
		"to_be_kept": 1,
		"to_be_upd":  []string{"testo"},
	}
	b := map[string]interface{}{
		"simple":    1,
		"to_add":    0, // to be added
		"to_delete": nil,
		"to_be_upd": []string{"testo", "pesto"},
	}

	r := MergeMaps(a, b)

	if v, ok := r["simple"]; !ok || v != 1 {
		t.Fatalf("simple value is not present or didn't change to be 1 after merge")
	}

	if v, ok := r["to_add"]; !ok || v != 0 {
		t.Fatalf("value to be added isn't added or value is not expected")
	}

	if _, ok := r["to_be_kept"]; !ok {
		t.Fatalf("value to be kept is deleted")
	}

	if _, ok := r["to_delete"]; ok {
		t.Fatalf("value to be deleted isn't deleted")
	}

	if v, ok := r["to_be_upd"]; !ok {
		t.Fatal("value to be updated been deleted")
	} else if l := len(v.([]string)); l != 2 {
		t.Fatalf("value was updated incorrectly, expected length to be 2, got: %d", l)
	}
}

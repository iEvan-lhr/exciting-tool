package orderedmap_test

import (
	"fmt"

	"github.com/iEvan-lhr/exciting-tool/orderedmap"
)

func ExampleMap() {
	values := orderedmap.New[string, int]()
	values.Set("first", 1)
	values.Set("second", 2)
	fmt.Println(values.Keys())
	// Output: [first second]
}

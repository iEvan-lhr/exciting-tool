package textutil_test

import (
	"fmt"

	"github.com/iEvan-lhr/exciting-tool/textutil"
)

func ExampleBetween() {
	value, ok := textutil.Between("<title>exciting-tool</title>", "<title>", "</title>")
	fmt.Println(value, ok)
	// Output: exciting-tool true
}

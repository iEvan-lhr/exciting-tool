package structuredtext_test

import (
	"fmt"

	"github.com/iEvan-lhr/exciting-tool/structuredtext"
)

func ExampleExtractJSON() {
	value, ok := structuredtext.ExtractJSON("result: {\"ok\":true}")
	fmt.Println(value, ok)
	// Output: {"ok":true} true
}

func ExampleMarkerTokenizer() {
	tokenizer, _ := structuredtext.NewMarkerTokenizer("(img:", ")")
	first, _ := tokenizer.Push("before (img")
	second, _ := tokenizer.Push(":file-1) after")
	last, _ := tokenizer.Flush()

	for _, tokens := range [][]structuredtext.Token{first, second, last} {
		for _, token := range tokens {
			fmt.Printf("%d:%q\n", token.Kind, token.Value)
		}
	}
	// Output:
	// 0:"before "
	// 1:"file-1"
	// 0:" after"
}

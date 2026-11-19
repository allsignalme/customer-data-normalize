package openai_test

import (
	"fmt"

	"github.com/allsignalme/customer-data-normalize/openai"
)

func ExampleHashIdentifier() {
	h, ok := openai.HashIdentifier(openai.IdentEmail, "  John.Doe@Example.COM ")
	fmt.Println(h, ok)
	// Output: 836f82db99121b3481011f16b49dfa5fbc714a0d1b1b9f784a1ebbbf5b39577f true
}

func ExampleInternationalPhone() {
	intl := openai.InternationalPhone("0983 714 965", "VN")
	fmt.Println(intl, openai.NormalizePhone(intl))
	// Output: +84983714965 84983714965
}

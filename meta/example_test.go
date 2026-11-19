package meta_test

import (
	"fmt"

	"github.com/allsignalme/customer-data-normalize/meta"
)

func ExampleHash() {
	em, ok := meta.Hash(meta.Email, " John_Smith@gmail.com ")
	fmt.Println(em, ok)
	_, ok = meta.Hash(meta.Email, "not-an-email")
	fmt.Println(ok)
	// Output:
	// 62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f true
	// false
}

func ExampleNormalize() {
	ph, _ := meta.Normalize(meta.Phone, "0912 345 678", meta.WithCountry("VN"))
	ct, _ := meta.Normalize(meta.City, "New York")
	db, _ := meta.Normalize(meta.DateOfBirth, "2/16/1997")
	fmt.Println(ph, ct, db)
	// Output: 84912345678 newyork 19970216
}

func ExampleShouldHash() {
	fmt.Println(meta.ShouldHash("em"), meta.ShouldHash("fbp"))
	// Output: true false
}

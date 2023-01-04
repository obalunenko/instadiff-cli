# image-diff
[![Go Report Card](https://goreportcard.com/badge/github.com/olegfedoseev/image-diff)](https://goreportcard.com/report/github.com/olegfedoseev/image-diff)

Allow you to calculate the difference between two images. Primarily for documents for now.


You give it first image, like that:

![test-only-text](testdata/test-only-text.png?raw=true "test-only-text")

And second one:

![test-text-number](testdata/test-text-number.png?raw=true "test-text-number")

And you get diff percent (6.25%) and visual diff:

![diff](testdata/diff.png?raw=true "diff")

# How to get and use

```
go get -u github.com/olegfedoseev/image-diff
```

And in you code:

```
import "github.com/olegfedoseev/image-diff"

diff, percent, err := diff.CompareFiles("test-only-text.png", "test-text-number.png")
if percent > 0.0 {
    fmt.Printf("images is different!")
}
```

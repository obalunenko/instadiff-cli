# gogroup

Check correctness of import groups in Go source, and fix files that have bad
grouping.

## Concept

Each project should choose a canonical import grouping. For example:

* First standard imports
* Then imports starting with "local/"
* Finally third party imports.

Groups should be separated by empty lines, and within each group imports should be sorted.

So this is allowed:

```go
// In a.go
import (
	"os"
	"testing"
	
	"local/bar"
	"local/foo"
	
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)
```

But this is not, because of an extra empty line, and `local/foo` being out of position.

```go
// In b.go
import (
	"os"
	
	"testing"
	
	"local/bar"
	
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"local/foo"
)
```

## Installation

Either install by running `go get github.com/vasi-stripe/gogroup/cmd/gogroup` or, 
if you have cloned the code, run `go install ./...` from the cloned dir.

## Usage

Check which files validate this order:

```bash
bash$ gogroup -order std,prefix=local/,other a.go b.go
b.go:6: Extra empty line inside import group at "testing"
bash$ echo $?
3
```

Fixup files to match this order:

```bash
bash$ gogroup -order std,prefix=local/,other -rewrite a.go b.go
Fixed b.go.
```

Then check git diff, to ensure that nothing broke. Now `b.go` should look like `a.go`.

## Support

The following import structures are currently supported:

1. A single import declaration, without parens:

	```go
	// Optional comment
	import "foo"
	```

1. A single import declaration with parens:

	```go
	// Optional comment
	import (
	  "something" // Optional comment
	  
	  // Optional comment
	  "another/thing"
	  "one/more/thing"
	  name "import/with/name"
	  . "dot/import"
	)
	```

All of these allow doc comments and named imports.
 
        
## TODO

* Write tests, check coverage
* Improve validation messages
* Figure out what to do with different structures:
	* Multiple import declarations. Are these allowed? Do we merge these together?

		```go
		import (
			"something"
		)
		import (
			"another/thing"
		)
		```   

	* Spaces at start/end of import declaration. Get rid of them?
	
		```go
		import (

			"something"

		)
		```
	
	* Random comments. Are these allowed? If we move things around, where do these comments end up?
	
		```go
		import (
			"something"
			
			// Random comment in the middle.
			
			// Normal doc comment, this is already ok.
			"another/thing"
			// Trailing comment.
		)
		
		// Comment in the middle of import declarations.
		
		// Normal doc comment, this is ok.
		import (
			"something/else"
		)
		```
	
	* `import "C"` statements. We should try hard to keep these separate from other imports.
	
		```go
		import (
			"something"
		)
		
		// Random comment.
		
		// #cgo CFLAGS: -DPNG_DEBUG=1
		// #cgo amd64 386 CFLAGS: -DX86=1
		// #cgo LDFLAGS: -lpng
		// #include <png.h>
		import "C"

		import (
			"another/thing"
		)
		```
		
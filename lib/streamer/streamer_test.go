package streamer

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestErrorToHeaders(t *testing.T) {
	for count := 1; count < 999; count++ {
		errStr := "err"
		err := errors.New(strings.Repeat(fmt.Sprintf("%s\n \n", errStr), count))
		k, v := errorToHeaders(err)
		if len(k) != len(v) || len(k) != count {
			t.Fatalf("%q*%d should return %d element slices", errStr, count, count)
		}
		header := fmt.Sprintf("%s0%d", defaultErrorHeader, count)
		if header != k[count-1] {
			t.Fatalf("%q*%d returned header slice %d element should be %q", errStr, count, count, header)
		}
		if v[count-1] != errStr {
			t.Fatalf("%q*%d should return slice of %q values", errStr, count, errStr)
		}
	}
}

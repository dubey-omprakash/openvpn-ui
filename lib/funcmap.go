package lib

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// AddFuncMaps .
func AddFuncMaps() {
	// safeid turns a certificate name into a string that is safe to use in an
	// HTML id and in a CSS/jQuery selector. Names such as "first.last" or
	// "user@example.com" contain characters that break "#id" selectors, which
	// stopped the certificate details popup from opening.
	_ = web.AddFuncMap("safeid", func(in string) string {
		var b strings.Builder
		changed := false
		for _, r := range in {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				b.WriteRune(r)
			} else {
				b.WriteByte('-')
				changed = true
			}
		}
		out := b.String()
		if changed {
			// keep "a.b" and "a-b" from colliding
			h := fnv.New32a()
			_, _ = h.Write([]byte(in))
			out = fmt.Sprintf("%s-%08x", out, h.Sum32())
		}
		return out
	})
	_ = web.AddFuncMap("field_error_message", func(v map[string]map[string]string, key string) map[string]string {
		if val, ok := v[key]; ok {
			return val
		}
		return make(map[string]string)
	})
	_ = web.AddFuncMap("field_error_exist", func(v map[string]map[string]string, key string) bool {
		if _, ok := v[key]; ok {
			return true
		}
		return false
	})
	_ = web.AddFuncMap("printkb", func(i interface{}) string {
		switch v := i.(type) {
		case uint64:
			return num2str(int64(i.(uint64)/1024), '\u00A0')
		case int64:
			return num2str(i.(int64)/1024, '\u00A0')
		default:
			logs.Error("Unknown type:", v)
		}
		return "Mapping error"
	})
	_ = web.AddFuncMap("printmb", func(i interface{}) string {
		switch v := i.(type) {
		case uint64:
			return num2str(int64(i.(uint64)/1024/1024), '\u00A0')
		case int64:
			return num2str(i.(int64)/1024/1024, '\u00A0')
		default:
			logs.Error("Unknown type:", v)
		}
		return "Mapping error"
	})
	_ = web.AddFuncMap("printmbold", func(i uint64) string {
		return num2str(int64(i/1024/1024), ' ')
	})
	_ = web.AddFuncMap("printgb", func(i uint64) string {
		return num2str(int64(i/1024/1024/1024), ' ')
	})
	_ = web.AddFuncMap("percent", func(x, y interface{}) string {
		//logs.Notice("Percent", x, y)
		zValue := "0"
		switch v := x.(type) {
		case string:
			logs.Error("Not implemented")
		case int32:
			if x.(int32) == 0 || y.(int32) == 0 {
				return zValue
			}
			a := float64(x.(int32))
			b := float64(y.(int32))
			return fmt.Sprintf("%d", int((a/b)*float64(100)))
		case int64:
			if x.(int64) == 0 || y.(int64) == 0 {
				return zValue
			}
			a := float64(x.(int64))
			b := float64(y.(int64))
			return fmt.Sprintf("%d", int((a/b)*float64(100)))
		case uint64:
			if x.(uint64) == 0 || y.(uint64) == 0 {
				return zValue
			}
			a := float64(x.(uint64))
			b := float64(y.(uint64))
			return fmt.Sprintf("%d", int((a/b)*float64(100)))
		default:
			logs.Error("Unknown type:", v)
		}
		return "Mapping error"
	})
}

func num2str(n int64, sep rune) string {
	s := strconv.FormatInt(n, 10)
	startOffset := 0
	var buff bytes.Buffer
	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}
	l := len(s)
	commaIndex := 3 - ((l - startOffset) % 3)
	if commaIndex == 3 {
		commaIndex = 0
	}
	for i := startOffset; i < l; i++ {
		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++
		buff.WriteByte(s[i])
	}
	return buff.String()
}

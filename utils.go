package gutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net"
	"net/http"
	"os"
	"reflect"
)

var (
	newLineB = []byte("\n")
	// the html codes for unescaping
	ltHex = []byte("\\u003c")
	lt    = []byte("<")

	gtHex = []byte("\\u003e")
	gt    = []byte(">")

	andHex = []byte("\\u0026")
	and    = []byte("&")
)

const ResponseBodyContextKey = "response.body"

func SubStr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

// ResponseText _
func ResponseText(ctx iris.Context, format string, args ...interface{}) (n int, err error) {
	defer func() {
		ctx.Values().Set(ResponseBodyContextKey, fmt.Sprintf(format, args))
	}()

	return ctx.Text(format, args)
}

// ResponseString _
func ResponseString(ctx iris.Context, body string) (n int, err error) {
	defer func() {
		ctx.Values().Set(ResponseBodyContextKey, body)
	}()

	return ctx.WriteString(body)
}

// ResponseJSON _
func ResponseJSON(ctx iris.Context, v interface{}, opts ...context.JSON) (n int, err error) {
	options := context.DefaultJSONOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	ctx.ContentType(context.ContentJSONHeaderValue)
	var (
		result   []byte
		optimize = ctx.Application().ConfigurationReadOnly().GetEnableOptimizations()
	)
	defer func() {
		ctx.Values().Set(ResponseBodyContextKey, string(result))
	}()

	if indent := options.Indent; indent != "" {
		marshalIndent := json.MarshalIndent
		if optimize {
			marshalIndent = jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent
		}

		result, err = marshalIndent(v, "", indent)
		result = append(result, newLineB...)
	} else {
		marshal := json.Marshal
		if optimize {
			marshal = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal
		}

		result, err = marshal(v)
	}

	if err != nil {
		ctx.Application().Logger().Debugf("JSON: %v", err)
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, err
	}

	if options.UnescapeHTML {
		result = bytes.Replace(result, ltHex, lt, -1)
		result = bytes.Replace(result, gtHex, gt, -1)
		result = bytes.Replace(result, andHex, and, -1)
	}

	if prefix := options.Prefix; prefix != "" {
		result = append([]byte(prefix), result...)
	}

	return ctx.Write(result)
}

// ResponseXML _
func ResponseXML(ctx iris.Context, v interface{}, opts ...context.XML) (n int, err error) {
	options := context.DefaultXMLOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	ctx.ContentType(context.ContentXMLHeaderValue)

	var (
		bf     bytes.Buffer
		result []byte
	)
	defer func() {
		ctx.Values().Set(ResponseBodyContextKey, string(bf.Bytes()))
	}()

	if prefix := options.Prefix; prefix != "" {
		n, err := bf.Write([]byte(prefix))
		if err != nil {
			return n, err
		}
	}

	if indent := options.Indent; indent != "" {
		result, err = xml.MarshalIndent(v, "", indent)
		if err == nil {
			result = append(result, newLineB...)
		}
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		ctx.Application().Logger().Debugf("XML: %v", err)
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, err
	}
	bf.Write(result)

	return ctx.Write(bf.Bytes())
}

func GetLocalIP() (string, error) {
	var ip string
	if ip = os.Getenv("POD_IP"); ip != "" {
		return ip, nil
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ip, err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", nil
}

func DoBatches(f func(interface{}), value interface{}, batchSize int) {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))

	switch reflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflectValue.Len(); i += batchSize {
			ends := i + batchSize
			if ends > reflectValue.Len() {
				ends = reflectValue.Len()
			}

			f(reflectValue.Slice(i, ends).Interface())
		}
	default:
		f(value)
	}
}

package gutils

import (
	"net"
	"os"
	"reflect"
)

func SubStr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
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

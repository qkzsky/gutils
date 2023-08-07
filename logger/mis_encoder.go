// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package logger

import (
	"fmt"
	"github.com/qkzsky/gutils/config"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type MisEncoder struct {
	zapcore.Encoder
}

func NewMisEncoder(cfg zapcore.EncoderConfig) *MisEncoder {
	return &MisEncoder{
		zapcore.NewJSONEncoder(cfg),
	}
}

func (t *MisEncoder) Clone() zapcore.Encoder {
	return &MisEncoder{t.Encoder.Clone()}
}

func (t *MisEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (buf *buffer.Buffer, err error) {
	line := bufferPool.Get()

	// [时间] [业务名] [主机名] [级别] [logID] 日志体
	// [2016-08-07 16:40:29] [my_service_name] [10.237.36.185] [NOTICE] [1306209370] {"appId":"my_service_name","clientIp":"127.0.0.1:24001","method":"HelloWorld","parent":977444679,"total":87}
	var hostName string
	hostName, err = os.Hostname()
	if err != nil {
		return
	}
	line.AppendString(fmt.Sprintf("[%s] [%s] [%s] [%s] [%s] ",
		ent.Time.Format(time.DateTime),
		config.AppName,
		hostName,
		ent.Level.CapitalString(),
		"0",
	))

	buf, err = t.Encoder.EncodeEntry(ent, fields)
	if err != nil {
		return
	}
	_, err = line.Write(buf.Bytes())
	if err != nil {
		return
	}

	return line, nil
}

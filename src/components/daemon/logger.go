package daemon

import (
	"encoding/gob"
	"io"
	"log/slog"
	"net/rpc"
	"os"
	"reflect"
)

var logger *slog.Logger

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
}

type LoggingServerCodec struct {
	dec *gob.Decoder
	enc *gob.Encoder
	c   io.Closer
}

func NewLoggingServerCodec(conn io.ReadWriteCloser) *LoggingServerCodec {
	return &LoggingServerCodec{
		dec: gob.NewDecoder(conn),
		enc: gob.NewEncoder(conn),
		c:   conn,
	}
}

func formatBody(body interface{}) slog.Value {
	if body == nil {
		return slog.StringValue("null")
	}
	v := reflect.ValueOf(body)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
		body = v.Interface()
	}
	return slog.AnyValue(body)
}

func (c *LoggingServerCodec) ReadRequestHeader(r *rpc.Request) error {
	err := c.dec.Decode(r)
	if err == nil {
		logger.Info("Received Request",
			"ServiceMethod", r.ServiceMethod,
		)
	}
	return err
}

func (c *LoggingServerCodec) ReadRequestBody(body interface{}) error {
	err := c.dec.Decode(body)
	if err == nil {
		logger.Info("Request Body",
			"Body", formatBody(body),
		)
	}
	return err
}

func (c *LoggingServerCodec) WriteResponse(r *rpc.Response, body interface{}) error {
	if err := c.enc.Encode(r); err != nil {
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		return err
	}
	logger.Info("Sent Response",
		"ServiceMethod", r.ServiceMethod,
		"Error", r.Error,
		"Body", formatBody(body),
	)
	return nil
}

func (c *LoggingServerCodec) Close() error {
	return c.c.Close()
}

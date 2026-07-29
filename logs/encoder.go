package logs

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type Encoder struct {
	zapcore.Encoder
}

func NewEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return &Encoder{Encoder: zapcore.NewJSONEncoder(cfg)}, nil
}

func (e *Encoder) Clone() zapcore.Encoder {
	return &Encoder{Encoder: e.Encoder.Clone()}
}

func (e *Encoder) EncodeEntry(log zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	fields = append(fields, zapcore.Field{Key: "severity", Type: zapcore.StringType, String: log.Level.CapitalString()})
	return e.Encoder.EncodeEntry(log, fields)
}

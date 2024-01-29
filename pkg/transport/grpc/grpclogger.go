package grpc

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

func replaceGRPCLogger(l *zerolog.Logger) {
	if l != nil {
		gl := &GrpcV2Logger{l}
		grpclog.SetLoggerV2(gl)
	}
}

var _ grpclog.LoggerV2 = &GrpcV2Logger{}

type GrpcV2Logger struct {
	L *zerolog.Logger
}

func (gl *GrpcV2Logger) Info(args ...interface{}) {
	gl.L.Debug().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Infoln(args ...interface{}) {
	gl.L.Debug().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Infof(format string, args ...interface{}) {
	gl.L.Debug().Msgf(format, args...)
}

func (gl *GrpcV2Logger) Warning(args ...interface{}) {
	gl.L.Warn().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Warningln(args ...interface{}) {
	gl.L.Warn().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Warningf(format string, args ...interface{}) {
	gl.L.Warn().Msgf(format, args...)
}

func (gl *GrpcV2Logger) Error(args ...interface{}) {
	gl.L.Error().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Errorln(args ...interface{}) {
	gl.L.Error().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Errorf(format string, args ...interface{}) {
	gl.L.Error().Msgf(format, args...)
}

func (gl *GrpcV2Logger) Fatal(args ...interface{}) {
	gl.L.Fatal().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Fatalln(args ...interface{}) {
	gl.L.Fatal().Msg(fmt.Sprint(args...))
}

func (gl *GrpcV2Logger) Fatalf(format string, args ...interface{}) {
	gl.L.Fatal().Msgf(format, args...)
}

func (gl *GrpcV2Logger) V(l int) bool {
	lvl := gl.L.GetLevel()
	switch lvl {
	case zerolog.TraceLevel, zerolog.DebugLevel:
		return true
	case zerolog.Disabled:
		return false
	}

	return int(lvl) <= l+1
}

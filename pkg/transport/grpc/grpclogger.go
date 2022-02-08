package grpc

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

func replaceGRPCLogger(l *zerolog.Logger) {
	if l != nil {
		gl := &grpcV2Logger{*l}
		grpclog.SetLoggerV2(gl)
	}
}

var _ grpclog.LoggerV2 = &grpcV2Logger{}

type grpcV2Logger struct {
	zerolog.Logger
}

func (gl *grpcV2Logger) Info(args ...interface{}) {
	gl.Logger.Debug().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Infoln(args ...interface{}) {
	gl.Logger.Debug().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Infof(format string, args ...interface{}) {
	gl.Logger.Debug().Msgf(format, args...)
}

func (gl *grpcV2Logger) Warning(args ...interface{}) {
	gl.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Warningln(args ...interface{}) {
	gl.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Warningf(format string, args ...interface{}) {
	gl.Logger.Warn().Msgf(format, args...)
}

func (gl *grpcV2Logger) Error(args ...interface{}) {
	gl.Logger.Error().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Errorln(args ...interface{}) {
	gl.Logger.Error().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Errorf(format string, args ...interface{}) {
	gl.Logger.Fatal().Msgf(format, args...)
}

func (gl *grpcV2Logger) Fatal(args ...interface{}) {
	gl.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Fatalln(args ...interface{}) {
	gl.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (gl *grpcV2Logger) Fatalf(format string, args ...interface{}) {
	gl.Logger.Fatal().Msgf(format, args...)
}

func (gl *grpcV2Logger) V(l int) bool {
	lvl := gl.Logger.GetLevel()
	switch lvl {
	case zerolog.TraceLevel, zerolog.DebugLevel:
		return true
	case zerolog.Disabled:
		return false
	}

	return int(lvl) <= l+1
}

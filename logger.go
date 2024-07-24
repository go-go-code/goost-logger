package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var instance *zap.Logger

func init() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	appRootPath := "."
	if os.Getenv("APP_ROOT") != "" {
		appRootPath = os.Getenv("APP_ROOT")
	}

	logFile := appRootPath + "/storage/logs/%Y%m%d.log"
	if os.Getenv("APP_NAME") != "" {
		logFile = appRootPath + "/storage/logs/" + os.Getenv("APP_NAME") + "-%Y%m%d.log"
	}
	rotator, err := rotatelogs.New(
		logFile,
		rotatelogs.WithMaxAge(10*24*time.Hour))
	if err != nil {
		panic(err)
	}

	writer := zapcore.AddSync(rotator)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	instance = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
}

func InfoF(f string, args ...any) {
	instance.Info(fmt.Sprintf(f, args...))
}

func Info(args ...any) {
	msg := format(args)
	if msg == "" {
		return
	}

	instance.Info(msg)
}

func DebugF(f string, args ...any) {
	instance.Debug(fmt.Sprintf(f, args...))
}

func Debug(args ...any) {
	msg := format(args)
	if msg == "" {
		return
	}

	instance.Debug(msg)
}

func ErrorF(f string, args ...any) {
	instance.Error(fmt.Sprintf(f, args...))
}

func Error(args ...any) {
	msg := format(args)
	if msg == "" {
		return
	}

	instance.Error(msg)
}

func WarnF(f string, args ...any) {
	instance.Warn(fmt.Sprintf(f, args...))
}

func Warn(args ...any) {
	msg := format(args)
	if msg == "" {
		return
	}

	instance.Warn(msg)
}

func format(args []any) string {
	if len(args) == 0 {
		return ""
	}

	if len(args) == 1 {
		return fmt.Sprintf("%+v", args[0])
	}

	b, _ := jsoniter.Marshal(args)

	return string(b)
}

func Access(c *gin.Context) {

	fields := []zap.Field{
		zap.String("type", "access"),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Int("status", c.Writer.Status()),
		zap.String("userAgent", c.Request.UserAgent()),
		zap.String("requestIP", c.ClientIP()),
		zap.String("requestID", requestid.Get(c)),
	}

	if referer := c.GetHeader("Referer"); referer != "" {
		fields = append(fields, zap.String("httpReferer", referer))
	}

	if sid := c.GetString("SID"); sid != "" {
		fields = append(fields, zap.String("clientID", sid))
	}

	if durations := c.GetString("request-durations"); durations != "" {
		fields = append(fields, zap.String("durations", durations))
	}

	if res := c.GetString("logger-response"); res != "" {
		fields = append(fields, zap.String("response", res))
	}

	instance.Info("", fields...)
}

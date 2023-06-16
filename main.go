package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/minhlong/go-aws-boilerplate/internal/handler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func main() {
	// Init log for debugging
	initLogger()

	// Declare lambda handle function
	lambda.Start(handler.HandleLambdaEvent)
}

func initLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	logFile, _ := os.OpenFile("logger.json5", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel

	// Log both file and console
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	zap.ReplaceGlobals(logger)
}

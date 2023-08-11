/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zap

import (
	"context"
	loggers "github.com/dynatrace/dynatrace-configuration-as-code-core/cmd/logr-test/log"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/cmd/logr-test/log/field"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/coordinate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

func customTimeEncoder(mode loggers.LogTimeMode) func(time.Time, zapcore.PrimitiveArrayEncoder) {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		layout := time.RFC3339
		if mode == loggers.LogTimeUTC {
			enc.AppendString(t.UTC().Format(layout))
		} else {
			enc.AppendString(t.Format(layout))
		}
	}
}

func New(logOptions loggers.LogOptions) (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder(logOptions.LogTimeMode)
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(levelMap[logOptions.LogLevel])

	var cores []zapcore.Core
	if logOptions.ConsoleLoggingJSON {
		consoleSyncer := zapcore.Lock(os.Stdout)
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), consoleSyncer, atomicLevel))
	} else {
		consoleSyncer := zapcore.Lock(os.Stderr)
		cores = append(cores, zapcore.NewCore(newFixedFieldsConsoleEncoder(), consoleSyncer, atomicLevel))
	}

	if logOptions.File != nil {
		fileSyncer := zapcore.Lock(zapcore.AddSync(logOptions.File))
		if logOptions.FileLoggingJSON {
			cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileSyncer, atomicLevel))
		} else {
			cores = append(cores, zapcore.NewCore(newFixedFieldsConsoleEncoder(), fileSyncer, atomicLevel))
		}
	}

	if logOptions.LogSpy != nil {
		spySyncer := zapcore.Lock(zapcore.AddSync(logOptions.LogSpy))
		if logOptions.ConsoleLoggingJSON {
			cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), spySyncer, atomicLevel))
		} else {
			cores = append(cores, zapcore.NewCore(newFixedFieldsConsoleEncoder(), spySyncer, atomicLevel))
		}

	}

	logger := zap.New(zapcore.NewTee(cores...))
	return logger, nil
}

// CtxKeyCoord context key used for contextual coordinate information
type CtxKeyCoord struct{}

// CtxKeyEnv context key used for contextual environment information
type CtxKeyEnv struct{}

// CtxValEnv context value used for contextual environment information
type CtxValEnv struct {
	Name  string
	Group string
}

// CtxGraphComponentId context key used for correlating logs that belong to deployment of a sub graph
type CtxGraphComponentId struct{}

// CtxValGraphComponentId context value used for correlating logs that belong to deployment of a sub graph
type CtxValGraphComponentId int

func WithCtxFields(loggr *zap.Logger, ctx context.Context) *zap.Logger {
	var f []zapcore.Field
	if c, ok := ctx.Value(CtxKeyCoord{}).(coordinate.Coordinate); ok {
		cF := field.Coordinate(c)
		f = append(f, zap.Reflect(cF.Key, cF.Value))
	}
	if e, ok := ctx.Value(CtxKeyEnv{}).(CtxValEnv); ok {
		eF := field.Environment(e.Name, e.Group)
		f = append(f, zap.Reflect(eF.Key, eF.Value))
	}

	if c, ok := ctx.Value(CtxGraphComponentId{}).(CtxValGraphComponentId); ok {
		f = append(f, zap.Int("gid", int(c)))
	}
	return loggr.With(f...)
}

var levelMap = map[loggers.LogLevel]zapcore.Level{
	loggers.LevelDebug: zapcore.DebugLevel,
	loggers.LevelInfo:  zapcore.InfoLevel,
	loggers.LevelWarn:  zapcore.WarnLevel,
	loggers.LevelError: zapcore.ErrorLevel,
	loggers.LevelFatal: zapcore.FatalLevel,
}

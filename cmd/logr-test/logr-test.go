package main

import (
	"context"
	"fmt"
	loggers "github.com/dynatrace/dynatrace-configuration-as-code-core/cmd/logr-test/log"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/cmd/logr-test/log/field"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/cmd/logr-test/log/zap"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/coordinate"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/afero"
	"io"
)

func main() {
	logfile, _ := afero.NewOsFs().Create("test.log")
	log, _ := zap.New(loggers.LogOptions{
		FileLoggingJSON:    true,
		ConsoleLoggingJSON: false,
		LogTimeMode:        loggers.LogTimeUTC,
		File:               logfile,
		LogLevel:           loggers.LevelDebug,
	})

	ctx := context.WithValue(context.TODO(), zap.CtxKeyCoord{}, coordinate.Coordinate{
		Project:  "test-project",
		Type:     "some-type",
		ConfigId: "awesome-config",
	})
	ctx = context.WithValue(ctx, zap.CtxKeyEnv{}, zap.CtxValEnv{Name: "test-env", Group: "test-group"})
	ctx = context.WithValue(ctx, zap.CtxGraphComponentId{}, zap.CtxValGraphComponentId(42))
	log = zap.WithCtxFields(log, ctx)

	rlog := zapr.NewLogger(log)

	c := rest.NewClient("http://google.com/rdn", rlog, rest.WithRequestRetrier(&rest.RequestRetrier{
		MaxRetries:      5,
		ShouldRetryFunc: rest.RetryIfNotSuccess,
	}))

	resp, err := c.GET(context.Background(), "")
	if err != nil {
		rlog.Error(err, "failed to get google homepage")
		return
	}

	b, err := io.ReadAll(resp.Body)

	rlog.Info(fmt.Sprintf("Received %v:\n%s[...]", resp.Status, b[0:100]))
}

func withZaprCtxFields(loggr logr.Logger, ctx context.Context) logr.Logger {
	var f []interface{}
	if c, ok := ctx.Value(zap.CtxKeyCoord{}).(coordinate.Coordinate); ok {
		cF := field.Coordinate(c)
		f = append(f, cF.Key, cF.Value)
	}
	if e, ok := ctx.Value(zap.CtxKeyEnv{}).(zap.CtxValEnv); ok {
		eF := field.Environment(e.Name, e.Group)
		f = append(f, eF.Key, eF.Value)
	}

	if c, ok := ctx.Value(zap.CtxGraphComponentId{}).(zap.CtxValGraphComponentId); ok {
		f = append(f, "gid", c)
	}
	return loggr.WithValues(f...)
}

package scoped

import (
	"context"

	"github.com/sirupsen/logrus"
)

var baseFormatter logrus.Formatter = &logrus.TextFormatter{
	DisableColors: false,
	FullTimestamp: true,
	ForceColors:   true,
}

func New(scope string, parentFormatter ...logrus.Formatter) *ScopedFormatter {
	var base logrus.Formatter = baseFormatter
	if len(parentFormatter) != 0 {
		base = parentFormatter[0]
	}
	return &ScopedFormatter{
		scope:           scope,
		parentFormatter: base,
	}
}

type ScopedFormatter struct {
	scope           string
	parentFormatter logrus.Formatter
}

func (f *ScopedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scope", f.scope)
	if entry.Data["scope"] == nil {
		entry.Data["scope"] = f.scope
	}
	return f.parentFormatter.Format(entry)
}

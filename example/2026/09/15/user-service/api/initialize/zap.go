package initialize

import "go.uber.org/zap"

func Zap() {
	l, err := zap.NewDevelopment()
	// l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
}

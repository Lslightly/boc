package boc

import "os"

func envHas(env string) bool {
	s := os.Getenv(env)
	return s != "" && s != "0" && s != "false"
}

var typecheck bool = envHas("TYPE_CHECK")
var debug bool = envHas("BOC_DEBUG")

package utils

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/fatih/color"
)

func Json(a ...any) {
	color.Yellow("%s spew json: \n", runFuncPos())
	for _, v := range a {
		d, _ := json.MarshalIndent(v, "", "\t")
		fmt.Println(string(d))
	}
}

func runFuncPos() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}

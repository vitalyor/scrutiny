package handler

import "github.com/analogj/scrutiny/webapp/backend/pkg/web/collect"

var collectRunner *collect.Runner

func SetCollectRunner(runner *collect.Runner) {
	collectRunner = runner
}

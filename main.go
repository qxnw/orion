package main

import (
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(hydra.WithPlatName("orion"),
		hydra.WithSystemName("logging"),
		hydra.WithServerTypes("api-rpc"),
		hydra.WithAutoCreateConf(),
		hydra.WithDebug())
	binding(app)
	app.Start()
}

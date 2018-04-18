package main

import (
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(hydra.WithPlatName("orion-2"),
		hydra.WithSystemName("logging"),
		hydra.WithServerTypes("api-rpc"),
		hydra.WithDebug())
	bind(app)
	app.Start()
}

/*
Copyright © 2024 Juha Ruotsalainen <juha.ruotsalainen@iki.fi>
*/
package main

import "mimfluxdb/cmd"

var (
	appVersion = "dev"
	appName    string
)

func main() {
	cmd.Execute(appName, appVersion)
}

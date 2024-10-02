package main

import (
    "EskgTgBotGo/app"
    "log"
)

func main() {
    if err := app.Run() ; err != nil {
        log.Fatal(err)
    }
}

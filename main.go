package main

import (
	"gotard/db"
	"gotard/server"
	"context"
	"fmt"
)

var ctx = context.Background()

func main() {

    err := db.InitDB(ctx, "users.db")
    if err != nil {
        fmt.Println("Error init'ing database:", err)
        return
    }
    defer db.DB.Close()

    server.StartServer()
}

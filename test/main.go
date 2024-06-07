package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gbrlmza/cpit"
)

func main() {
	ctx := context.Background()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	cpit.SetDefaultBaseUrl(os.Getenv("CPIT_BASEURL"))
	cpit.SetDefaultApiKey(os.Getenv("CPIT_APIKEY"))
	cpit.SetDefaultDebugMode(true)

	data := map[string]interface{}{
		"id":     "81b0d1d3666134637a00037c",
		"title":  "mytitle",
		"number": time.Now().Unix(),
	}

	err := cpit.UpsertItem(ctx, "test", cpit.WithBody(data))
	if err != nil {
		log.Fatal(err)
	}
}

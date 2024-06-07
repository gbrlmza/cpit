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
		"_id":    "906668493839374024000373",
		"title":  "UPDATED Title",
		"number": time.Now().Unix(),
	}

	err := cpit.UpsertItem(ctx, "test", cpit.WithData(data))
	if err != nil {
		log.Fatal(err)
	}
}

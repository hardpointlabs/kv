package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/hardpointlabs/invar/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logger := zerolog.New(os.Stdout)
	adapter := &BadgerZerologAdapter{Logger: logger}

	opts := badger.DefaultOptions("/tmp/badger")
	opts.Logger = adapter
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open badger database")
	}

	defer db.Close()

	go func() {
		log.Info().Msg("starting pprof server on localhost:6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("pprof server failed")
		}
	}()

	redis.Serve(db)
}

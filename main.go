// Command mct is the v2 Master's Closet server. Boot order is intentional and
// resolves the chicken-and-egg between config and the database (the db path +
// encryption key live in config, but config itself lives in the db): the JSON
// file passed as the first argument is parsed only to learn those bootstrap
// fields, the db is opened, and then the live config is loaded from (or seeded
// into) the db.
//
// The v1 server still exists, isolated under ./legacy (build: go build ./legacy).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/fingerprint"
	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/paths"
	"github.com/0187773933/MastersCloset/v2/server"
	"github.com/0187773933/MastersCloset/v2/user"
	bleve "github.com/blevesearch/bleve/v2"
	bolt "github.com/boltdb/bolt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: mct <config.json>")
		os.Exit(1)
	}

	// 1. Parse the seed file for bootstrap-only fields.
	seed, err := config.ParseSeedFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if seed.BoltDBPath == "" {
		seed.BoltDBPath = "mct.db"
	}
	if seed.BleveSearchPath == "" {
		seed.BleveSearchPath = "mct.bleve"
	}

	// 2. Resolve per-machine data paths the v1 way (mct_<fingerprint>.db under
	//    ~/.config/mct) so an existing install's data is found automatically.
	//    These become the authoritative bootstrap paths.
	fp := fingerprint.Short()
	dbPath := paths.ResolveFingerprinted(seed.BoltDBPath, fp, false)
	blevePath := paths.ResolveFingerprinted(seed.BleveSearchPath, fp, true)
	seed.BoltDBPath = dbPath
	seed.BleveSearchPath = blevePath
	seed.FingerPrint = fp

	// 3. Open the database.
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		fmt.Printf("could not open db at %s: %v\n", dbPath, err)
		os.Exit(1)
	}

	// 3. Load or seed the live config from the db.
	cfg := config.NewManager(db)
	seeded, err := cfg.LoadOrSeed(seed)
	if err != nil {
		fmt.Printf("config load/seed failed: %v\n", err)
		os.Exit(1)
	}
	snap := cfg.Snapshot()

	// 4. Logger (now that we know the timezone).
	logger.Init(os.Getenv("LOG_LEVEL"), snap.TimeZone)
	log := logger.GetLogger()
	if seeded {
		log.Info(fmt.Sprintf("seeded config into db from %s", os.Args[1]))
	} else {
		log.Info("loaded live config from db")
	}
	log.Info(fmt.Sprintf("machine fingerprint: %s", fp))
	log.Info(fmt.Sprintf("db: %s", dbPath))

	// 5. Open or create the Bleve search index.
	index, err := bleve.Open(blevePath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Info(fmt.Sprintf("creating bleve index at %s", blevePath))
		index, err = bleve.New(blevePath, bleve.NewIndexMapping())
	}
	if err != nil {
		log.Info(fmt.Sprintf("search index unavailable (%v); continuing without search", err))
		index = nil
	}

	// 6. Wire the store and server.
	store := user.NewStore(db, cfg, index)
	srv := server.New(cfg, db, index, store)

	// 7. Graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-c
		fmt.Println("\r- shutting down")
		srv.Shutdown()
		if index != nil {
			index.Close()
		}
		db.Close()
		logger.Close()
		cancel()
		os.Exit(0)
	}()
	_ = ctx

	log.Info("Hola , Christ Lives")
	if err := srv.Start(); err != nil {
		log.Info(fmt.Sprintf("server stopped: %v", err))
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/crypto"
	"github.com/wadeling/origin-check/internal/relay"
	"github.com/wadeling/origin-check/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	enc, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		log.Fatal(err)
	}

	path := "config/seeds/relays.yaml"
	seeds := relay.DefaultSeeds()
	if _, err := os.Stat(path); err == nil {
		seeds, err = relay.LoadSeedFile(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := relay.Seed(ctx, st, enc, seeds); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeded relays successfully")
}

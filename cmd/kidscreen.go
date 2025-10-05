package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/albertb/kidscreen/internal"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to get user home directory:", err)
	}
	defaultConfigPath := filepath.Join(home, ".config", "kidscreen", "config.yaml")

	configPath := flag.String("config", defaultConfigPath, "path to the config file")
	dev := flag.Bool("dev", false, "whether to keep the webserver running for dev")
	addr := flag.String("addr", ":9999", "the address the webserver should listen on in dev mode")
	fake := flag.Bool("fake", false, "whether to generate fake weather and calendar data in dev mode")
	img := flag.String("img", "screen.png", "the path to save the rendered image")

	flag.Parse()

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal("failed to open config file:", err)
	}
	defer configFile.Close()

	cfg, err := internal.ReadConfig(configFile)
	if err != nil {
		log.Fatal("failed to read config file:", err)
	}

	if err := internal.Run(cfg, *dev, *fake, *img, *addr); err != nil {
		log.Fatal("failed to render:", err)
	}
}

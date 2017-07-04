package main

import (
	"log"

	"github.com/mfojtik/jenkins-prometheus/pkg/collector"
)

func main() {
	log.Print("Starting manager ...")
	collector.Run()
}

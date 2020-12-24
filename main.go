package main

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/service"
	"time"
)

func main() {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Minute),
	)
	defer cancel()

	// fire up the simulation
	service.Init()
	service.Run(ctx)
	defer service.Done()

	// wait for content expire
	select {
	case <-ctx.Done():
		return
	}
}

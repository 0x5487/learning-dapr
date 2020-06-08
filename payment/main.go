package main

import (
	"time"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"github.com/jasonsoft/napnap"
)

type RegisterActor struct {
	Entities                []string `json:"entities"`
	ActorIdleTimeout        string   `json:"actorIdleTimeout"`
	ActorScanInterval       string   `json:"actorScanInterval"`
	DrainOngoingCallTimeout string   `json:"drainOngoingCallTimeout"`
	DrainRebalancedActors   bool     `json:"drainRebalancedActors"`
}

func main() {
	// use console handler to log all level logs
	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)

	router := napnap.NewRouter()

	router.Get("/healthz", func(c *napnap.Context) {
		// now := time.Now()
		// log.Debugf("healthz...%v", now)
		c.SetStatus(200)
	})

	router.Get("/dapr/config", func(c *napnap.Context) {

		regActor := RegisterActor{
			Entities:                []string{"payment"},
			ActorIdleTimeout:        "1h",
			ActorScanInterval:       "30s",
			DrainOngoingCallTimeout: "10s",
			DrainRebalancedActors:   true,
		}

		c.JSON(200, regActor)
	})

	router.Delete("/actors/payment/:actorID", func(c *napnap.Context) {
		actorID := c.Param("actorID")
		log.Debugf("payment: delete actor: param_id: %s", actorID)
	})

	router.Put("/actors/payment/:actorID/method/resvered-amount", func(c *napnap.Context) {
		actorID := c.Param("actorID")

		log.Debugf("payment: resvered-amount: param_id: %s", actorID)

		time.Sleep(2 * time.Second)

		log.Infof("payment: resvered-amount success: param_id: %s", actorID)

		c.SetStatus(500)
	})

	router.Put("/actors/payment/:actorID/method/resvered-amount-failed", func(c *napnap.Context) {
		actorID := c.Param("actorID")

		log.Debugf("payment: resvered-amount: param_id: %s", actorID)

		c.SetStatus(500)
	})

	nap := napnap.New()
	nap.Use(router)
	httpEngine := napnap.NewHttpEngine("127.0.0.1:3002")
	nap.Run(httpEngine)
}

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

type ReminderPayload struct {
	Data    interface{} `json:"data"`
	DueTime string      `json:"dueTime"`
	Period  string      `json:"period"`
}

type DataContent struct {
	Content string `json:"content"`
}

type Message struct {
	Specversion string      `json:"specversion"`
	Data        interface{} `json:"data"`
}

func main() {
	// use console handler to log all level logs
	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)

	router := napnap.NewRouter()

	router.Get("/healthz", func(c *napnap.Context) {
		//now := time.Now()
		//log.Debugf("healthz...%v", now)
		c.SetStatus(200)
	})

	router.Get("/dapr/config", func(c *napnap.Context) {

		regActor := RegisterActor{
			Entities:                []string{"order"},
			ActorIdleTimeout:        "1h",
			ActorScanInterval:       "30s",
			DrainOngoingCallTimeout: "10s",
			DrainRebalancedActors:   true,
		}

		c.JSON(200, regActor)
	})

	router.Delete("/actors/order/:actorID", func(c *napnap.Context) {
		actorID := c.Param("actorID")
		log.Debugf("delete actor: param_id: %s", actorID)
	})

	router.Put("/actors/order/:actorID/method/create-order", func(c *napnap.Context) {
		actorID := c.Param("actorID")

		log.Debugf("order creating: param_id: %s", actorID)

		time.Sleep(2 * time.Second)

		log.Debugf("order created: param_id: %s", actorID)

		c.SetStatus(200)
	})

	router.Put("/actors/order/:actorID/method/remind/check-sms", func(c *napnap.Context) {
		actorID := c.Param("actorID")
		utcNow := time.Now().UTC()
		log.Debugf("reminder: %s, now: %v", actorID, utcNow)

		c.SetStatus(500)
		return

		c.SetStatus(200)
	})

	// cron
	router.Options("/mycron", func(c *napnap.Context) {
		c.RespHeader("Allow", "POST")
		c.RespHeader("Content-Type", "application/javascript")
		c.SetStatus(200)
	})

	router.Post("/mycron", func(c *napnap.Context) {
		log.Debugf("cron was triggered %s", time.Now())
		time.Sleep(5 * time.Second)
		c.SetStatus(200)
	})

	nap := napnap.New()
	nap.Use(router)
	httpEngine := napnap.NewHttpEngine("127.0.0.1:3001")
	nap.Run(httpEngine)
}

package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jasonsoft/learning-dapr/internal/pkg/dapr"
	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"github.com/jasonsoft/napnap"
	"github.com/jasonsoft/request"
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

type Order struct {
	ID              string `json:"id"`
	MerchantOrderNO string `json:"merchant_order_no"`
	Amount          int64  `json:"amount"`
}

type Message struct {
	Specversion string `json:"specversion"`
	Order       Order  `json:"data"`
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
			Entities:                []string{"order-orchestration"},
			ActorIdleTimeout:        "1h",
			ActorScanInterval:       "30s",
			DrainOngoingCallTimeout: "10s",
			DrainRebalancedActors:   true,
		}

		c.JSON(200, regActor)
	})

	// subscribe topics
	router.Get("/dapr/subscribe", func(c *napnap.Context) {

		topics := []string{"create-order-request"}
		c.JSON(200, topics)
	})

	router.Post("/v1.0/orders", func(c *napnap.Context) {
		// publish create order event
		log.Debug("==begin create order ===")

		url := "http://localhost:3500/v1.0/publish/create-order-request"

		// "c8f06e98-4c79-4081-b27e-b88ec034143f"
		order := Order{
			ID:              uuid.New().String(),
			MerchantOrderNO: "abc",
			Amount:          100,
		}

		resp, err := request.POST(url).
			SendJSON(order).
			End()

		if err != nil {
			log.Err(err).Error("can't send to dapr: for creat-order-request")
			return
		}

		if resp.OK == false {
			log.Debugf("creat-order-request failed %s", resp.String())
			c.String(500, resp.String())
			return
		}

		// query tx state
		for {
			val, err := dapr.ActorState("order-orchestration", order.ID, "status")
			if err != nil {
				if errors.Is(err, dapr.ErrStateKeyNotFound) {
					log.Warnf("key:status was not found")
				} else {
					log.Err(err).Error("get status failed")
					c.String(500, err.Error())
					return
				}
			}
			if strings.EqualFold(val, "\"order_created\"") {
				c.String(200, fmt.Sprintf("order has been created: %s", order.ID))
				return
			}
			time.Sleep(1 * time.Second)
			log.Debug("check tx status...")
		}
	})

	router.Post("/create-order-request", func(c *napnap.Context) {
		msg := Message{}
		err := c.BindJSON(&msg)
		if err != nil {
			log.Err(err).Warn("bind json err")
			return
		}
		log.Str("spec", msg.Specversion).
			Debugf("create-order-req: merchant_order_no: %s", msg.Order.MerchantOrderNO)

		// invoke order orachestion actor
		url := fmt.Sprintf("http://localhost:3500/v1.0/actors/order-orchestration/%s/method/create-order-tx", msg.Order.ID)
		resp, err := request.PUT(url).
			SendJSON(nil).
			End()

		if err != nil {
			log.Err(err).Error("can't send to dapr: for creat-order-tx")
			return
		}

		if resp.OK == false {
			log.Debugf("create order tx  failed %s", resp.String())
			c.String(500, resp.String())
			return
		}

		log.Debug("create order tx completed")
		c.SetStatus(200)
	})

	// order orchestration actor
	router.Put("/actors/order-orchestration/:actorID/method/create-order-tx", func(c *napnap.Context) {
		ctx := c.StdContext()
		actorID := c.Param("actorID")
		log.Debugf("tx: order creating: order_id: %s", actorID)

		orch, err := NewOrchestration(actorID)
		if err != nil {
			c.String(500, err.Error())
			return
		}

		switch orch.StateMachine.State() {
		case "initial":
			err = orch.StateMachine.Trigger(ctx, "order_created")
			if err != nil {
				c.String(500, err.Error())
				return
			}
		case "order_created":
			err = orch.StateMachine.Trigger(ctx, "order_paid")
			if err != nil {
				// change state to payment failed
				c.String(500, err.Error())
				return
			}

		case "order_paid":
			err = orch.StateMachine.Trigger(ctx, "order_created")
			if err != nil {
				c.String(500, err.Error())
				return
			}
		}

		// get state first
		val, err := dapr.ActorState("order-orchestration", actorID, "status")
		if err != nil {
			if errors.Is(err, dapr.ErrStateKeyNotFound) {
				log.Warnf("key:status was not found")
			} else {
				log.Err(err).Error("get status failed")
				c.String(500, err.Error())
				return
			}
		}
		log.Debugf("state: key:status value: %s", val)

		if strings.EqualFold(val, "\"order_created\"") {
			log.Info("order has been created")
			c.String(200, "order has been created")
			return
		}
		log.Debugf("val not match: %s, %s", val, "order_created")

		// invoke order actor
		url := fmt.Sprintf("http://localhost:3500/v1.0/actors/order/%s/method/create-order", actorID)
		resp, err := request.PUT(url).
			SendJSON(nil).
			End()

		if err != nil {
			log.Err(err).Error("can't send to dapr: can't invoke order actor create-order")
			return
		}

		if resp.OK == false {
			log.Debugf("invoke order actor create-order  failed %s", resp.String())
			c.String(500, resp.String())
			return
		}

		err = dapr.SaveActorState("order-orchestration", actorID, "status", "order_created")
		if err != nil {
			log.Err(err).Warn("invoke order actor create-order completed")
		}

		// invoke payment actor
		// url = fmt.Sprintf("http://localhost:3500/v1.0/actors/payment/%s/method/resvered-amount-failed", actorID)
		// resp, err = request.PUT(url).
		// 	SendJSON(nil).
		// 	End()

		// if err != nil {
		// 	log.Err(err).Error("can't send to dapr: can't invoke payment actor; method: resvered-amount")
		// 	_ = dapr.SaveActorState("order-orchestration", actorID, "status", "payment_failed")
		// 	c.String(500, resp.String())
		// 	return
		// }

		// if resp.OK == false {
		// 	log.Debugf("invoke payment actor, method: resvered-amount failed %s", resp.String())
		// 	c.String(500, resp.String())
		// 	return
		// }

		err = dapr.SaveActorState("order-orchestration", actorID, "resp", "order_created")
		if err != nil {
			log.Err(err).Warn("invoke order actor create-order completed")
		}

		log.Info("order orchestration completed")
		c.SetStatus(200)
	})

	nap := napnap.New()
	nap.Use(router)
	httpEngine := napnap.NewHttpEngine("127.0.0.1:3000")
	nap.Run(httpEngine)
}

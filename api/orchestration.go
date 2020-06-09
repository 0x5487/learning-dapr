package main

import (
	"context"

	"github.com/jasonsoft/fsm"
	"github.com/jasonsoft/learning-dapr/internal/pkg/dapr"
)

type Orchestration struct {
	StateMachine *fsm.StateMachine
}

func NewOrchestration(id string) (*Orchestration, error) {
	orch := Orchestration{}

	sm := fsm.New("initial").
		Transition("order_created").From("initial").To("order_created").
		Before(func(ctx context.Context, e *fsm.Event) error {
			return orch.OrderCreated(ctx, e)
		}).
		Transition("order_paid").From("order_created").To("order_paid").
		Before(func(ctx context.Context, e *fsm.Event) error {
			return orch.OrderPaid(ctx, e)
		}).
		Transition("payment_failed").From("order_created").To("payment_failed").
		Transition("order_failed").From("payment_failed").To("order_failed").
		Before(func(ctx context.Context, e *fsm.Event) error {
			return orch.PaymentFailed(ctx, e)
		}).
		StateMachine()

	status, err := dapr.ActorState("order-orchestration", id, "status")
	if err != nil {
		return nil, err
	}

	if status != "" {
		sm.SetState(status)
	}

	orch.StateMachine = sm
	return &orch, nil
}

func (orch *Orchestration) OrderCreated(ctx context.Context, e *fsm.Event) error {
	return nil
}

func (orch *Orchestration) OrderPaid(ctx context.Context, e *fsm.Event) error {
	return nil
}

func (orch *Orchestration) PaymentFailed(ctx context.Context, e *fsm.Event) error {
	return nil
}

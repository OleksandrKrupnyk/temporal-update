package updates

import (
	"math/rand"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	UPDUpper = "MakeUpper"
	UPDStop  = "MakeStop"
	SIGStop  = "Final"
)

func Workflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow started.", "StartTime", workflow.Now(ctx))
	active := true

	if err := workflow.SetUpdateHandler(ctx, UPDUpper, func(ctx workflow.Context, greet string) (string, error) {
		delay := rand.Intn(5)
		workflow.Sleep(ctx, time.Duration(delay)*time.Second)
		return strings.ToUpper(greet), nil

	}); err != nil {
		logger.Error("SetUpdateHandler", "Error", err)
	}

	if err := workflow.SetUpdateHandler(ctx, UPDStop, func(ctx workflow.Context) error {
		active = false
		delay := rand.Intn(5)
		workflow.Sleep(ctx, time.Duration(delay)*time.Second)
		logger.Info("Active set stage false")
		return nil

	}); err != nil {
		logger.Error("SetUpdateHandler", "Error", err)
	}
	stopCh := workflow.GetSignalChannel(ctx, SIGStop)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(stopCh, func(c workflow.ReceiveChannel, more bool) {
		active = false
	})
	for {
		selector.Select(ctx)
		if !active {
			break
		}
	}

	if err := workflow.Await(ctx, func() bool {
		return workflow.AllHandlersFinished(ctx)
	}); err != nil {
		logger.Error("AllHandlersFinished", "Error", err)
	}
	logger.Info("Workflow finished.", "FinishTime", workflow.Now(ctx))

	return nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	updates "work-temporal-updates"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /item", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		c, err := client.Dial(client.Options{
			HostPort: client.DefaultHostPort,
		})
		if err != nil {
			slog.Error("Unable to create client", slog.String("error", err.Error()))
			return
		}
		defer c.Close()
		workFlowID := "sd1"

		describer, err := c.DescribeWorkflowExecution(ctx, workFlowID, "")
		if errors.Is(err, &serviceerror.NotFound{}) || describer.GetWorkflowExecutionInfo().GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
			_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
				ID:        workFlowID,
				TaskQueue: "default",
			}, updates.Workflow)
			if err != nil {
				slog.Error("Unable execute workflow", slog.String("error", err.Error()))
				return
			}
		} else if err != nil {
			slog.Error("Unable to describe workflow", slog.String("error", err.Error()))
			return
		}
		slog.Info("update workflow")
		updateHandle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workFlowID,
			RunID:        "",
			UpdateName:   updates.UPDUpper,
			WaitForStage: client.WorkflowUpdateStageAccepted,
			Args:         []interface{}{"payload string to upper"},
		})
		if err != nil {
			slog.Error("Unable to update workflow", slog.String("error", err.Error()))
			return
		}
		var stage string
		if err = updateHandle.Get(ctx, &stage); err != nil {
			log.Fatalln("Unable to get workflow stage", err)
		}
		slog.Info(fmt.Sprintf("workflow stage: %s", stage))
	})

	mux.HandleFunc("PUT /item", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		d := time.Now().Add(14 * time.Second)
		ctx, done := context.WithDeadline(r.Context(), d)
		defer done()
		c, err := client.Dial(client.Options{
			HostPort: client.DefaultHostPort,
		})
		if err != nil {
			slog.Error("Unable to create client", err)
			return
		}
		defer c.Close()

		workFlowID := "sd1"
		stopHandle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workFlowID,
			RunID:        "",
			UpdateName:   updates.UPDStop,
			WaitForStage: client.WorkflowUpdateStageAccepted,
			Args:         []interface{}{},
		})
		if err != nil {
			slog.Error("Unable to update workflow", slog.String("error", err.Error()))
			return
		}
		if err = stopHandle.Get(ctx, nil); err != nil {
			slog.Error("Unable to get workflow stage", slog.String("error", err.Error()))
			return
		}
	})

	mux.HandleFunc("DELETE /item", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()
		c, err := client.Dial(client.Options{
			HostPort: client.DefaultHostPort,
		})
		if err != nil {
			log.Fatalln("Unable to create client", err)
		}
		defer c.Close()
		workFlowID := "sd1"
		if err = c.SignalWorkflow(ctx, workFlowID, "", updates.SIGStop, nil); err != nil {
			slog.Error("Unable to signal workflow", slog.String("error", err.Error()))
			return
		}
	})

	mux.HandleFunc("GET /item", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		fmt.Fprint(w, "got path\n")
		return
	})
	log.Println("starting server")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		slog.Error("Unable to start server", slog.String("error", err.Error()))
		return
	}
}

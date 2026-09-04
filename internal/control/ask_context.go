package control

import (
	"context"
	"strings"

	"reasonix/internal/event"
)

// Ask emits an AskRequest and blocks until AnswerQuestion answers it. Unlike
// tool approvals, Ask is never bypassed by YOLO because it represents a genuine
// user decision.
func (c *Controller) Ask(ctx context.Context, questions []event.AskQuestion) ([]event.AskAnswer, error) {
	return c.askWithContext(ctx, "", questions)
}

// AskWithContext retains reviewable Markdown across frontend reconnects.
func (c *Controller) AskWithContext(ctx context.Context, reviewContext string, questions []event.AskQuestion) ([]event.AskAnswer, error) {
	return c.askWithContext(ctx, strings.TrimSpace(reviewContext), questions)
}

func (c *Controller) askWithContext(ctx context.Context, reviewContext string, questions []event.AskQuestion) ([]event.AskAnswer, error) {
	id, reply := c.approval.registerAsk(reviewContext, questions)
	if !c.lockPromptFor(ctx, "question") {
		c.approval.cancelAsk(id)
		return nil, ctx.Err()
	}
	defer c.approval.promptMu.Unlock()

	c.approval.promptEmitMu.Lock()
	if err := event.EmitChecked(c.sink, event.Event{Kind: event.AskRequest, ItemID: id, Ask: event.Ask{ID: id, Context: reviewContext, Questions: questions}}); err != nil {
		c.approval.promptEmitMu.Unlock()
		c.approval.cancelAsk(id)
		return nil, err
	}
	c.approval.markAskEmitted(id)
	c.approval.promptEmitMu.Unlock()

	waitCtx, cancelWait := c.approval.waitContext(ctx)
	defer cancelWait()
	select {
	case ans := <-reply:
		return ans, nil
	case <-waitCtx.Done():
		c.approval.cancelAsk(id)
		return nil, waitCtx.Err()
	}
}

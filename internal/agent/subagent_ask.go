package agent

import "context"

func inheritCallAsker(sub *Agent, ctx context.Context, enabled bool) *Agent {
	if !enabled {
		return sub
	}
	_, _, asker, ok := CallContext(ctx)
	if ok && asker != nil {
		sub.SetAsker(asker)
	}
	return sub
}

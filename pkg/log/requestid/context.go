package requestid

import "context"

var (
	DefaultRequestIDKey = "x-request-id"
)

func FromContext(ctx context.Context) string {
	id, ok := ctx.Value(DefaultRequestIDKey).(string)
	if !ok {
		return ""
	}
	return id
}

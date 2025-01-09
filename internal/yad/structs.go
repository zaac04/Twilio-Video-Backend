package yad

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Trace struct {
	TraceID    string            `json:"trace_id"`
	Spans      []*Span           `json:"spans"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Duration   time.Duration     `json:"duration"`
	Status     Status            `json:"status"`
	Attributes map[string]string `json:"attributes"`
}

type Span struct {
	TraceID       string  `json:"trace_id"`
	SpanID        string  `json:"span_id"`
	ParentSpanID  string  `json:"parent_span_id"`
	OperationName string  `json:"operation_name"`
	Events        []Event `json:"events"`
}

type Event struct {
	Timestamp  time.Time              `json:"timestamp"`
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func StartNewTrace() *Trace {
	return &Trace{
		TraceID:   uuid.NewString(),
		StartTime: time.Now(),
	}
}

func (t *Trace) CreateSpan(functionName string) *Span {
	var ParentSpanID string

	SpanID := uuid.NewString()
	if len(t.Spans) != 0 {
		ParentSpanID = t.Spans[len(t.Spans)-1].SpanID
	} else {
		ParentSpanID = SpanID
	}

	return &Span{
		TraceID:       t.TraceID,
		ParentSpanID:  ParentSpanID,
		SpanID:        SpanID,
		OperationName: functionName,
	}
}

func (s *Span) AddEvent(Key string, Value interface{}) {
	event := Event{
		Timestamp: time.Now(),
		Attributes: map[string]interface{}{
			Key: Value,
		},
	}
	s.Events = append(s.Events, event)
}

func (t *Trace) AddSpan(span *Span) {
	t.Spans = append(t.Spans, span)
}

func GetTrace(ctx context.Context) *Trace {
	return ctx.Value(ReqCtxKey("trace")).(ReqContext).Trace
}

type ReqContext struct {
	ReqID     string
	UserId    string
	SessionId string
	Role      string
	Trace     *Trace
	Span      *Span
}

type ReqCtxKey string

func (t *Trace) AddTraceToCtx(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), ReqCtxKey("trace"), ReqContext{ReqID: t.TraceID, Trace: t})
	r = r.WithContext(ctx)
	return r
}

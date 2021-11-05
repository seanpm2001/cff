package example

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/cff"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// Request TODO
type Request struct {
	LDAPGroup string
}

// Response TODO
type Response struct {
	MessageIDs []string
}

type fooHandler struct {
	mgr    *ManagerRepository
	users  *UserRepository
	ses    *SESClient
	scope  tally.Scope
	logger *zap.Logger
}

func (h *fooHandler) HandleFoo(ctx context.Context, req *Request) (*Response, error) {
	var res *Response
	err := func(
		ctx context.Context,
		emitter cff.Emitter,
		v1 *Request,
	) (err error) {
		var (
			flowInfo = &cff.FlowInfo{
				Name:   "HandleFoo",
				File:   "go.uber.org/cff/examples/magic.go",
				Line:   32,
				Column: 9,
			}
			flowEmitter = emitter.FlowInit(flowInfo)

			schedInfo = &cff.SchedulerInfo{
				Name:      flowInfo.Name,
				Directive: cff.FlowDirective,
				File:      flowInfo.File,
				Line:      flowInfo.Line,
				Column:    flowInfo.Column,
			}

			// possibly unused
			_ = flowInfo
		)

		startTime := time.Now()
		defer func() { flowEmitter.FlowDone(ctx, time.Since(startTime)) }()

		schedEmitter := emitter.SchedulerInit(schedInfo)

		sched := cff.BeginFlow(8, schedEmitter)

		type task struct {
			emitter cff.TaskEmitter
			ran     cff.AtomicBool
			run     func(context.Context) error
			job     *cff.ScheduledJob
		}

		type predicate struct {
			ran cff.AtomicBool
			run func(context.Context) error
			job *cff.ScheduledJob
		}

		var tasks []*task
		defer func() {
			for _, t := range tasks {
				if !t.ran.Load() {
					t.emitter.TaskSkipped(ctx, err)
				}
			}
		}()

		// go.uber.org/cff/examples/magic.go:41:4
		var (
			v2 *GetManagerRequest
			v3 *ListUsersRequest
		)
		task0 := new(task)
		task0.emitter = cff.NopTaskEmitter()
		task0.run = func(ctx context.Context) (err error) {
			taskEmitter := task0.emitter
			startTime := time.Now()
			defer func() {
				if task0.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanic(ctx, recovered)
					err = fmt.Errorf("task panic: %v", recovered)
				}
			}()

			defer task0.ran.Store(true)

			v2, v3 = func(req *Request) (*GetManagerRequest, *ListUsersRequest) {
				return &GetManagerRequest{
						LDAPGroup: req.LDAPGroup,
					}, &ListUsersRequest{
						LDAPGroup: req.LDAPGroup,
					}
			}(v1)

			taskEmitter.TaskSuccess(ctx)

			return
		}

		task0.job = sched.Enqueue(ctx, cff.Job{
			Run: task0.run,
		})
		tasks = append(tasks, task0)

		// go.uber.org/cff/examples/magic.go:49:4
		var (
			v4 *GetManagerResponse
		)
		task1 := new(task)
		task1.emitter = cff.NopTaskEmitter()
		task1.run = func(ctx context.Context) (err error) {
			taskEmitter := task1.emitter
			startTime := time.Now()
			defer func() {
				if task1.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanic(ctx, recovered)
					err = fmt.Errorf("task panic: %v", recovered)
				}
			}()

			defer task1.ran.Store(true)

			v4, err = h.mgr.Get(v2)

			if err != nil {
				taskEmitter.TaskError(ctx, err)
				return err
			} else {
				taskEmitter.TaskSuccess(ctx)
			}

			return
		}

		task1.job = sched.Enqueue(ctx, cff.Job{
			Run: task1.run,
			Dependencies: []*cff.ScheduledJob{
				task0.job,
			},
		})
		tasks = append(tasks, task1)

		// go.uber.org/cff/examples/magic.go:61:4
		var (
			v5 *ListUsersResponse
		)
		task4 := new(task)
		task4.emitter = emitter.TaskInit(
			&cff.TaskInfo{
				Name:   "FormSendEmailRequest",
				File:   "go.uber.org/cff/examples/magic.go",
				Line:   61,
				Column: 4,
			},
			&cff.DirectiveInfo{
				Name:      flowInfo.Name,
				Directive: cff.FlowDirective,
				File:      flowInfo.File,
				Line:      flowInfo.Line,
				Column:    flowInfo.Column,
			},
		)
		task4.run = func(ctx context.Context) (err error) {
			taskEmitter := task4.emitter
			startTime := time.Now()
			defer func() {
				if task4.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanicRecovered(ctx, recovered)
					v5, err = &ListUsersResponse{}, nil
				}
			}()

			defer task4.ran.Store(true)

			v5, err = h.users.List(v3)

			if err != nil {
				taskEmitter.TaskErrorRecovered(ctx, err)
				v5, err = &ListUsersResponse{}, nil
			} else {
				taskEmitter.TaskSuccess(ctx)
			}

			return
		}

		task4.job = sched.Enqueue(ctx, cff.Job{
			Run: task4.run,
			Dependencies: []*cff.ScheduledJob{
				task0.job,
			},
		})
		tasks = append(tasks, task4)

		// go.uber.org/cff/examples/magic.go:74:4
		var p0 bool
		pred1 := new(predicate)
		pred1.run = func(ctx context.Context) (err error) {
			p0 = func(req *GetManagerRequest) bool {
				return req.LDAPGroup != "everyone"
			}(v2)
			return nil
		}

		pred1.job = sched.Enqueue(ctx, cff.Job{
			Run: pred1.run,
			Dependencies: []*cff.ScheduledJob{
				task0.job,
			},
		})

		// go.uber.org/cff/examples/magic.go:66:4
		var (
			v6 []*SendEmailRequest
		)
		task5 := new(task)
		task5.emitter = emitter.TaskInit(
			&cff.TaskInfo{
				Name:   "FormSendEmailRequest",
				File:   "go.uber.org/cff/examples/magic.go",
				Line:   66,
				Column: 4,
			},
			&cff.DirectiveInfo{
				Name:      flowInfo.Name,
				Directive: cff.FlowDirective,
				File:      flowInfo.File,
				Line:      flowInfo.Line,
				Column:    flowInfo.Column,
			},
		)
		task5.run = func(ctx context.Context) (err error) {
			taskEmitter := task5.emitter
			startTime := time.Now()
			defer func() {
				if task5.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanic(ctx, recovered)
					err = fmt.Errorf("task panic: %v", recovered)
				}
			}()

			if !p0 {
				return nil
			}

			defer task5.ran.Store(true)

			v6 = func(mgr *GetManagerResponse, users *ListUsersResponse) []*SendEmailRequest {
				var reqs []*SendEmailRequest
				for _, u := range users.Emails {
					reqs = append(reqs, &SendEmailRequest{Address: u})
				}
				return reqs
			}(v4, v5)

			taskEmitter.TaskSuccess(ctx)

			return
		}

		task5.job = sched.Enqueue(ctx, cff.Job{
			Run: task5.run,
			Dependencies: []*cff.ScheduledJob{
				task1.job,
				task4.job,
				pred1.job,
			},
		})
		tasks = append(tasks, task5)

		// go.uber.org/cff/examples/magic.go:50:12
		var (
			v7 []*SendEmailResponse
		)
		task2 := new(task)
		task2.emitter = cff.NopTaskEmitter()
		task2.run = func(ctx context.Context) (err error) {
			taskEmitter := task2.emitter
			startTime := time.Now()
			defer func() {
				if task2.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanic(ctx, recovered)
					err = fmt.Errorf("task panic: %v", recovered)
				}
			}()

			defer task2.ran.Store(true)

			v7, err = h.ses.BatchSendEmail(v6)

			if err != nil {
				taskEmitter.TaskError(ctx, err)
				return err
			} else {
				taskEmitter.TaskSuccess(ctx)
			}

			return
		}

		task2.job = sched.Enqueue(ctx, cff.Job{
			Run: task2.run,
			Dependencies: []*cff.ScheduledJob{
				task5.job,
			},
		})
		tasks = append(tasks, task2)

		// go.uber.org/cff/examples/magic.go:52:4
		var (
			v8 *Response
		)
		task3 := new(task)
		task3.emitter = cff.NopTaskEmitter()
		task3.run = func(ctx context.Context) (err error) {
			taskEmitter := task3.emitter
			startTime := time.Now()
			defer func() {
				if task3.ran.Load() {
					taskEmitter.TaskDone(ctx, time.Since(startTime))
				}
			}()

			defer func() {
				recovered := recover()
				if recovered != nil {
					taskEmitter.TaskPanic(ctx, recovered)
					err = fmt.Errorf("task panic: %v", recovered)
				}
			}()

			defer task3.ran.Store(true)

			v8 = func(responses []*SendEmailResponse) *Response {
				var r Response
				for _, res := range responses {
					r.MessageIDs = append(r.MessageIDs, res.MessageID)
				}
				return &r
			}(v7)

			taskEmitter.TaskSuccess(ctx)

			return
		}

		task3.job = sched.Enqueue(ctx, cff.Job{
			Run: task3.run,
			Dependencies: []*cff.ScheduledJob{
				task2.job,
			},
		})
		tasks = append(tasks, task3)

		if err := sched.Wait(ctx); err != nil {
			flowEmitter.FlowError(ctx, err)
			return err
		}

		*(&res) = v8 // *go.uber.org/cff/examples.Response

		flowEmitter.FlowSuccess(ctx)
		return nil
	}(
		ctx,
		cff.EmitterStack(cff.TallyEmitter(h.scope), cff.LogEmitter(h.logger)),
		req,
	)

	err = func(
		ctx context.Context,
		emitter cff.Emitter,
	) (err error) {
		var (
			parallelInfo = &cff.ParallelInfo{
				Name:   "SendParallel",
				File:   "go.uber.org/cff/examples/magic.go",
				Line:   82,
				Column: 8,
			}
			parallelEmitter = emitter.ParallelInit(parallelInfo)

			schedInfo = &cff.SchedulerInfo{
				Name:      parallelInfo.Name,
				Directive: cff.ParallelDirective,
				File:      parallelInfo.File,
				Line:      parallelInfo.Line,
				Column:    parallelInfo.Column,
			}

			// possibly unused
			_ = parallelInfo
		)

		startTime := time.Now()
		defer func() { parallelEmitter.ParallelDone(ctx, time.Since(startTime)) }()

		schedEmitter := emitter.SchedulerInit(schedInfo)

		sched := cff.BeginFlow(2, schedEmitter)

		type task struct {
			run func(context.Context) error
		}

		// go.uber.org/cff/examples/magic.go:89:4
		task6 := new(task)
		task6.run = func(ctx context.Context) (err error) {
			defer func() {
				recovered := recover()
				if recovered != nil {
					err = fmt.Errorf("parallel function panic: %v", recovered)
				}
			}()

			err = func(_ context.Context) error {
				return SendMessage()
			}(ctx)
			return
		}

		sched.Enqueue(ctx, cff.Job{
			Run: task6.run,
		})

		// go.uber.org/cff/examples/magic.go:92:4
		task7 := new(task)
		task7.run = func(ctx context.Context) (err error) {
			defer func() {
				recovered := recover()
				if recovered != nil {
					err = fmt.Errorf("parallel function panic: %v", recovered)
				}
			}()

			err = SendMessage()
			return
		}

		sched.Enqueue(ctx, cff.Job{
			Run: task7.run,
		})

		// go.uber.org/cff/examples/magic.go:95:4
		task8 := new(task)
		task8.run = func(ctx context.Context) (err error) {
			defer func() {
				recovered := recover()
				if recovered != nil {
					err = fmt.Errorf("parallel function panic: %v", recovered)
				}
			}()

			err = func() error {
				return SendMessage()
			}()
			return
		}

		sched.Enqueue(ctx, cff.Job{
			Run: task8.run,
		})

		if err := sched.Wait(ctx); err != nil {
			parallelEmitter.ParallelError(ctx, err)
			return err
		}
		parallelEmitter.ParallelSuccess(ctx)
		return nil
	}(
		ctx,
		cff.EmitterStack(cff.TallyEmitter(h.scope), cff.LogEmitter(h.logger)),
	)
	return res, err
}

// ManagerRepository TODO
type ManagerRepository struct{}

// GetManagerRequest TODO
type GetManagerRequest struct {
	LDAPGroup string
}

// GetManagerResponse TODO
type GetManagerResponse struct {
	Email string
}

// Get TODO
func (*ManagerRepository) Get(req *GetManagerRequest) (*GetManagerResponse, error) {
	return &GetManagerResponse{Email: "boss@example.com"}, nil
}

// UserRepository TODO
type UserRepository struct{}

// ListUsersRequest TODO
type ListUsersRequest struct {
	LDAPGroup string
}

// ListUsersResponse TODO
type ListUsersResponse struct {
	Emails []string
}

// List TODO
func (*UserRepository) List(req *ListUsersRequest) (*ListUsersResponse, error) {
	return &ListUsersResponse{
		Emails: []string{"a@example.com", "b@example.com"},
	}, nil
}

// SESClient TODO
type SESClient struct{}

// SendEmailRequest TODO
type SendEmailRequest struct {
	Address string
}

// SendEmailResponse TODO
type SendEmailResponse struct {
	MessageID string
}

// BatchSendEmail TODO
func (*SESClient) BatchSendEmail(req []*SendEmailRequest) ([]*SendEmailResponse, error) {
	res := make([]*SendEmailResponse, len(req))
	for i := range req {
		res[i] = &SendEmailResponse{MessageID: strconv.Itoa(i)}
	}
	return res, nil
}

// SendMessage returns nil error.
func SendMessage() error {
	return nil
}

package kapbit_test

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/kapbit/kapbit-go"
	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/executor"
	"github.com/kapbit/kapbit-go/repository/mem"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

type TestCase struct {
	Name                   string
	Setup                  Setup
	Prepare                func(t *testing.T, s evs.EventStore) error
	Items                  []Item
	Want                   Want
	IdempotencyCheckParams []workflow.Params
}

type Setup struct {
	NodeID           string
	Defs             []workflow.Definition
	Codec            evs.Codec
	PartitionCount   int
	PayloadSizeLimit int
	Opts             []kapbit.SetOption
	Parallel         bool
}

type Item struct {
	Params     workflow.Params
	WantResult workflow.Result
	WantError  error
}

type Want []WantItem

type WantItem struct {
	Wait   time.Duration
	Chpts  []int64
	Events [][]evnt.Event
}

func runTestCase(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		// handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// 	Level: slog.LevelDebug,
		// })
		// slog.SetDefault(slog.New(handler))

		repo := mem.New(tc.Setup.PartitionCount, tc.Setup.PayloadSizeLimit)

		opts := []kapbit.SetOption{
			kapbit.WithLogger(test.NopLogger),
			kapbit.WithNodeID(tc.Setup.NodeID),
			kapbit.WithTimeProvider(evnt.TimeProviderFn(
				func() int64 { return 0 },
			)),
		}

		if len(tc.Setup.Opts) > 0 {
			opts = append(opts, tc.Setup.Opts...)
		}

		if tc.Prepare != nil {
			err := tc.Prepare(t, evs.New(tc.Setup.NodeID, tc.Setup.Codec, repo))
			assertfatal.EqualError(t, err, nil)
		}

		kbit, err := kapbit.New(t.Context(), tc.Setup.Defs, tc.Setup.Codec,
			repo, opts...,
		)
		asserterror.EqualError(t, err, nil)

		wg := &sync.WaitGroup{}
		for _, item := range tc.Items {
			wg.Add(1)
			go func() {
				result, err := kbit.ExecWorkflow(item.Params.ID, item.Params.Type, item.Params.Input)
				asserterror.EqualError(t, err, item.WantError)
				asserterror.EqualDeep(t, result, item.WantResult)
				wg.Done()
			}()
			if !tc.Setup.Parallel {
				wg.Wait()
			}
		}
		wg.Wait()

		for _, item := range tc.Want {
			time.Sleep(item.Wait)
			if tc.Setup.Parallel {
				verifyEvents(t, item, repo, tc.Setup.Codec)
			} else {
				verifySequentialEvents(t, item, repo, tc.Setup.Codec)
			}
		}
		verifyIdempotencyCheck(t, tc, kbit)
	})
}

func verifySequentialEvents(t *testing.T, item WantItem,
	storage *mem.Repository, codec evs.Codec,
) {
	for partition := 0; partition < storage.PartitionCount(); partition++ {
		i := 0
		err := storage.ScanPartition(t.Context(), partition,
			func(chpt int64) error {
				asserterror.Equal(t, chpt, item.Chpts[partition])
				return nil
			},
			func(offset int64, beforeChpt bool, record evs.EventRecord) error {
				event, err := codec.Decode(record)
				asserterror.EqualError(t, err, nil)
				asserterror.EqualDeep(t, event, item.Events[partition][i])
				i++
				return nil
			})
		asserterror.EqualError(t, err, nil)
		asserterror.Equal(t, i, len(item.Events[partition]))
	}
}

func verifyEvents(t *testing.T, item WantItem, storage *mem.Repository,
	codec evs.Codec,
) {
	wantMap := makeWantMap(item.Events)
	for partition := 0; partition < storage.PartitionCount(); partition++ {
		i := 0
		err := storage.ScanPartition(t.Context(), partition,
			func(chpt int64) error {
				if len(item.Chpts) > 0 {
					asserterror.Equal(t, chpt, item.Chpts[partition])
				} else {
					fmt.Println(chpt)
				}
				return nil
			},
			func(offset int64, beforeChpt bool, record evs.EventRecord) error {
				event, err := codec.Decode(record)
				assertfatal.EqualError(t, err, nil)

				key := generateEventKey(event)
				if v, pst := wantMap[partition][key]; pst {
					if v.Count == 1 {
						delete(wantMap[partition], key)
					} else {
						v.Count--
						wantMap[partition][key] = v
					}
					asserterror.EqualDeep(t, event, v.Event)
				} else {
					t.Errorf("unexpected event %s, partition=%d", key, partition)
				}
				i++
				return nil
			})
		asserterror.EqualError(t, err, nil)
		asserterror.Equal(t, i, len(item.Events[partition]))
	}
}

func verifyIdempotencyCheck(t *testing.T, tc TestCase, kbit *kapbit.Kapbit) {
	for _, p := range tc.IdempotencyCheckParams {
		_, err := kbit.ExecWorkflow(p.ID, p.Type, p.Input)
		asserterror.EqualError(t, err,
			kapbit.NewKapbitError(executor.ErrIdempotencyViolation))
	}
}

func makeWantMap(want [][]evnt.Event) []map[string]struct {
	Event evnt.Event
	Count int
} {
	m := []map[string]struct {
		Event evnt.Event
		Count int
	}{}
	for range want {
		m = append(m, make(map[string]struct {
			Event evnt.Event
			Count int
		}))
	}

	for i, events := range want {
		for _, event := range events {
			key := generateEventKey(event)

			if v, pst := m[i][key]; pst {
				v.Count++
				m[i][key] = v
				continue
			}

			m[i][key] = struct {
				Event evnt.Event
				Count int
			}{Event: event, Count: 1}

		}
	}
	return m
}

func generateEventKey(event evnt.Event) string {
	switch e := event.(type) {
	case evnt.StepOutcomeEvent:
		return string(e.Key) + ":" + reflect.TypeFor[evnt.StepOutcomeEvent]().String() +
			":" + e.OutcomeType.String() + ":" + strconv.Itoa(int(e.OutcomeSeq))

	case evnt.WorkflowCreatedEvent:
		return string(e.Key) + ":" + reflect.TypeFor[evnt.WorkflowCreatedEvent]().String()
	case *evnt.WorkflowRejectedEvent:
		return string(e.Key) + ":" + reflect.TypeFor[evnt.WorkflowRejectedEvent]().String()
	case *evnt.DeadLetterEvent:
		return string(e.Key) + ":" + reflect.TypeFor[evnt.DeadLetterEvent]().String()
	case evnt.WorkflowResultEvent:
		return string(e.Key) + ":" + reflect.TypeFor[evnt.WorkflowResultEvent]().String()
	case evnt.ActiveWriterEvent:
		return string(e.NodeID) + ":" + reflect.TypeFor[evnt.ActiveWriterEvent]().String()
	default:
		panic(fmt.Sprintf("unsupported event type %T", e))
	}
}

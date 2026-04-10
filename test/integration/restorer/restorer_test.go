package restorer_test

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	mock "github.com/kapbit/kapbit-go/test/mock"
	utils "github.com/kapbit/kapbit-go/test/restorer"
	wflimpl "github.com/kapbit/kapbit-go/workflow/impl"
)

const NodeID = "node_id"

var (
	logger     = test.NopLogger
	eventTools = &evnt.Kit{}
	factory    = wflimpl.NewFactory(defs, eventTools, logger)
)

func TestRestorer(t *testing.T) {
	// TODO Add skip outcomes.
	store := mock.NewEventStoreMock()
	registerLoadRecent(store, -1, historyP1) // partition 1
	registerLoadRecent(store, 1, historyP2)  // partition 2
	registerLoadRecent(store, 6, historyP3)  // partition 3
	store.RegisterNPartitionCount(2, func() int { return 3 })

	tc := utils.TestCase{
		Name: "Should be able to restore a runtime",
		Config: utils.TestConfig{
			NodeID: NodeID,
			Factory: factory,
			Store:  store,
			Opts: []rtm.SetOption{
				rtm.WithCheckpointSize(10),
				rtm.WithIdempotencyDepth(3),
				rtm.WithLogger(logger),
			},
		},
		Want: utils.Want{
			CounterValue: 2,

			Refs: refs(),

			IdempotencyWindowStr: "support.FIFOSet cap=9 len=9 " +
				"[wfl-3 wfl-1 wfl-5 wfl-2 wfl-6 wfl-7 wfl-9 wfl-10 wfl-11]",

			ChptmanStr: "checkpoint.Manager partitions=3 chptSize=10 " +
				"[p0=\"chpt:{-1 -1} [{2 4},{2 4}]\" p1=\"chpt:{-1 1} [{0 2},{2 8}]\" p2=\"chpt:{-1 6}\"]",

			TrackerStr: "runtime.PositionTracker partitions=3 [p0:NextN=3 p1:NextN=3 p2:NextN=0] " +
				"tracked=[wfl-1:{0, 1, 2} wfl-3:{0, 0, 1}]",
		},
	}
	utils.RunTest(t, tc)
}

func registerLoadRecent(store mock.EventStoreMock, chpt int64,
	history []EventRecord,
) {
	store.RegisterLoadRecent(
		func(ctx context.Context, partition, workflowsCount int,
			chptFn evs.CheckpointCallback, eventFn evs.EventCallback,
		) (err error) {
			if err = chptFn(chpt); err != nil {
				return
			}
			for _, record := range history {
				err = eventFn(record.offset, record.beforeChpt, record.event)
				if err != nil {
					return
				}
			}
			return
		},
	)
}

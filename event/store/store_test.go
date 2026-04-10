package store_test

import (
	"testing"

	test "github.com/kapbit/kapbit-go/test/store"
)

func TestStoreSaveEvent(t *testing.T) {
	for _, tc := range []test.SaveEventTestCase{
		test.SaveEventEncodeErrorTestCase(),
		test.SaveEventRepositoryErrorTestCase(),
		test.SaveEventSuccessTestCase(),
	} {
		test.RunSaveEventTest(t, tc)
	}
}

func TestStoreLoadRecent(t *testing.T) {
	for _, tc := range []test.LoadRecentTestCase{
		test.LoadRecentRepositoryErrorTestCase(),
		test.LoadRecentDecodeErrorTestCase(),
		test.LoadRecentSuccessTestCase(),
	} {
		test.RunLoadRecentTest(t, tc)
	}
}

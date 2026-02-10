package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hunter-io/faktory/util"
	"github.com/stretchr/testify/assert"
)

func TestBasicQueueOps(t *testing.T) {
	withRedis(t, "queue", func(t *testing.T, store Store) {

		t.Run("Push", func(t *testing.T) {
			store.Flush()
			q, err := store.GetQueue("default")
			assert.NoError(t, err)

			assert.EqualValues(t, 0, q.Size())

			data, err := q.Pop()
			assert.NoError(t, err)
			assert.Nil(t, data)

			err = q.Push(5, []byte("hello"))
			assert.NoError(t, err)
			assert.EqualValues(t, 1, q.Size())

			err = q.Push(5, []byte("world"))
			assert.NoError(t, err)
			assert.EqualValues(t, 2, q.Size())

			values := [][]byte{
				[]byte("world"),
				[]byte("hello"),
			}
			q.Each(func(idx int, value []byte) error {
				assert.Equal(t, values[idx], value)
				return nil
			})

			data, err = q.Pop()
			assert.NoError(t, err)
			assert.Equal(t, []byte("hello"), data)
			assert.EqualValues(t, 1, q.Size())

			cnt, err := q.Clear()
			assert.NoError(t, err)
			assert.EqualValues(t, 0, cnt)
			assert.EqualValues(t, 0, q.Size())

			// valid names:
			_, err = store.GetQueue("A-Za-z0-9_.-")
			assert.NoError(t, err)
			_, err = store.GetQueue("-")
			assert.NoError(t, err)
			_, err = store.GetQueue("A")
			assert.NoError(t, err)
			_, err = store.GetQueue("a")
			assert.NoError(t, err)

			// invalid names:
			_, err = store.GetQueue("default?page=1")
			assert.Error(t, err)
			_, err = store.GetQueue("user@example.com")
			assert.Error(t, err)
			_, err = store.GetQueue("c&c")
			assert.Error(t, err)
			_, err = store.GetQueue("priority|high")
			assert.Error(t, err)
			_, err = store.GetQueue("")
			assert.Error(t, err)
		})

		t.Run("heavy", func(t *testing.T) {
			store.Flush()
			q, err := store.GetQueue("default")
			assert.NoError(t, err)

			assert.EqualValues(t, 0, q.Size())
			err = q.Push(5, []byte("first"))
			assert.NoError(t, err)
			n := 5000
			// Push N jobs to queue
			// Get Size() each time
			for i := 0; i < n; i++ {
				_, data := fakeJob()
				err = q.Push(5, data)
				assert.NoError(t, err)
				assert.EqualValues(t, i+2, q.Size())
			}

			err = q.Push(5, []byte("last"))
			assert.NoError(t, err)
			assert.EqualValues(t, n+2, q.Size())

			q, err = store.GetQueue("default")
			assert.NoError(t, err)

			// Pop N jobs from queue
			// Get Size() each time
			assert.EqualValues(t, n+2, q.Size())
			data, err := q.Pop()
			assert.NoError(t, err)
			assert.Equal(t, []byte("first"), data)
			for i := 0; i < n; i++ {
				_, err := q.Pop()
				assert.NoError(t, err)
				assert.EqualValues(t, n-i, q.Size())
			}
			data, err = q.Pop()
			assert.NoError(t, err)
			assert.Equal(t, []byte("last"), data)
			assert.EqualValues(t, 0, q.Size())

			data, err = q.Pop()
			assert.NoError(t, err)
			assert.Nil(t, data)
		})

		t.Run("threaded", func(t *testing.T) {
			store.Flush()
			q, err := store.GetQueue("default")
			assert.NoError(t, err)

			tcnt := 5
			n := 1000

			var wg sync.WaitGroup
			for i := 0; i < tcnt; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					pushAndPop(t, n, q)
				}()
			}

			wg.Wait()
			assert.EqualValues(t, 0, counter)
			assert.EqualValues(t, 0, q.Size())

			q.Each(func(idx int, v []byte) error {
				atomic.AddInt64(&counter, 1)
				//log.Println(string(k), string(v))
				return nil
			})
			assert.EqualValues(t, 0, counter)
		})
	})
}

var (
	counter int64
)

func pushAndPop(t *testing.T, n int, q Queue) {
	for i := 0; i < n; i++ {
		_, data := fakeJob()
		err := q.Push(5, data)
		assert.NoError(t, err)
		atomic.AddInt64(&counter, 1)
	}

	for i := 0; i < n; i++ {
		value, err := q.Pop()
		assert.NoError(t, err)
		assert.NotNil(t, value)
		atomic.AddInt64(&counter, -1)
	}
}

func TestBPopRespectsContext(t *testing.T) {
	withRedis(t, "bpop-ctx", func(t *testing.T, store Store) {
		store.Flush()
		q, err := store.GetQueue("default")
		assert.NoError(t, err)

		// A cancelled context should return immediately with the context error,
		// rather than blocking for the 2-second BRPop timeout.
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		start := time.Now()
		data, err := q.BPop(ctx)
		elapsed := time.Since(start)

		assert.Nil(t, data)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.True(t, elapsed < 500*time.Millisecond, "BPop should return immediately on cancelled context, took: %v", elapsed)
	})
}

func TestEachQueueConcurrentAccess(t *testing.T) {
	withRedis(t, "eachqueue", func(t *testing.T, store Store) {
		store.Flush()

		// Create some initial queues
		for i := 0; i < 5; i++ {
			_, err := store.GetQueue(fmt.Sprintf("queue-%d", i))
			assert.NoError(t, err)
		}

		// Concurrently iterate and add queues — this should not panic or race.
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				store.EachQueue(func(q Queue) {
					_ = q.Name()
				})
			}(i)
		}
		for i := 5; i < 15; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				_, _ = store.GetQueue(fmt.Sprintf("queue-%d", n))
			}(i)
		}
		wg.Wait()

		// Verify all queues are accessible
		count := 0
		store.EachQueue(func(q Queue) {
			count++
		})
		assert.True(t, count >= 5, "expected at least 5 queues, got %d", count)
	})
}

func fakeJob() (string, []byte) {
	return fakeJobWithPriority(5)
}

func fakeJobWithPriority(priority uint64) (string, []byte) {
	jid := util.RandomJid()
	nows := util.Nows()
	return jid, []byte(fmt.Sprintf(`{"jid":"%s","created_at":"%s","priority":%d,"queue":"default","args":[1,2,3],"class":"SomeWorker"}`, jid, nows, priority))
}

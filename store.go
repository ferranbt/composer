package runner

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/assert"
	gproto "google.golang.org/protobuf/proto"
)

var (
	allocsBucket = []byte("allocs")
	allocKey     = []byte("alloc")
	taskStateKey = []byte("task-state")
)

func taskKey(name string) []byte {
	return []byte("task-" + name)
}

func newInmemStore(t *testing.T) *BoltdbStore {
	tmpDir := t.TempDir()

	state, err := NewBoltdbStore(filepath.Join(tmpDir, "my.db"))
	assert.NoError(t, err)

	return state
}

type BoltdbStore struct {
	db *bolt.DB
}

func NewBoltdbStore(path string) (*BoltdbStore, error) {
	db, err := bolt.Open(path, 0755, &bolt.Options{})
	if err != nil {
		return nil, err
	}

	return NewBoltdbStoreWithDB(db)
}

func NewBoltdbStoreWithDB(db *bolt.DB) (*BoltdbStore, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			allocsBucket,
		}
		for _, bkt := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bkt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s := &BoltdbStore{
		db: db,
	}
	return s, nil
}

func (s *BoltdbStore) PutProject(a *proto.Project) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allocsBkt := tx.Bucket(allocsBucket)

		bkt, err := allocsBkt.CreateBucketIfNotExists([]byte(a.Name))
		if err != nil {
			return err
		}
		if err := dbPut(bkt, allocKey, a); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *BoltdbStore) GetProjects() ([]*proto.Project, error) {
	allocs := []*proto.Project{}
	s.db.View(func(tx *bolt.Tx) error {
		allocsBkt := tx.Bucket(allocsBucket)

		c := allocsBkt.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			allocBkt := allocsBkt.Bucket(k)

			alloc := proto.Project{}
			if err := dbGet(allocBkt, allocKey, &alloc); err != nil {
				return err
			}
			allocs = append(allocs, &alloc)
		}
		return nil
	})
	return allocs, nil
}

func (s *BoltdbStore) GetTasks(allocID string) ([]string, error) {
	tasks := []string{}

	err := s.db.View(func(tx *bolt.Tx) error {
		allocsBkt := tx.Bucket(allocsBucket)

		allocBkt := allocsBkt.Bucket([]byte(allocID))
		if allocBkt == nil {
			return fmt.Errorf("alloc '%s' not found", allocID)
		}

		allocBkt.ForEach(func(k, v []byte) error {
			tasks = append(tasks, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task state %v", err)
	}

	return tasks, nil
}

func (s *BoltdbStore) GetTaskState(allocID, taskName string) (*proto.TaskState, error) {
	state := proto.TaskState{}

	err := s.db.View(func(tx *bolt.Tx) error {
		allocsBkt := tx.Bucket(allocsBucket)

		allocBkt := allocsBkt.Bucket([]byte(allocID))
		if allocBkt == nil {
			return fmt.Errorf("alloc '%s' not found", allocID)
		}

		taskBkt := allocBkt.Bucket(taskKey(taskName))
		if taskBkt == nil {
			return fmt.Errorf("task '%s' not found", taskName)
		}
		if err := dbGet(taskBkt, taskStateKey, &state); err != nil {
			return fmt.Errorf("failed to get task state '%s' '%s'", allocID, taskName)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task state %v", err)
	}

	return &state, nil
}

func (s *BoltdbStore) PutTaskState(allocID string, taskName string, state *proto.TaskState) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allocsBkt := tx.Bucket(allocsBucket)

		allocBkt := allocsBkt.Bucket([]byte(allocID))
		if allocBkt == nil {
			return fmt.Errorf("alloc '%s' not found", allocID)
		}

		taskBkt, err := allocBkt.CreateBucketIfNotExists(taskKey(taskName))
		if err != nil {
			return err
		}
		if err := dbPut(taskBkt, taskStateKey, state); err != nil {
			return fmt.Errorf("failed to get task state '%s' '%s'", allocID, taskName)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to put task state: %v", err)
	}
	return nil
}

func dbPut(bkt *bolt.Bucket, key []byte, obj gproto.Message) error {
	raw, err := gproto.Marshal(obj)
	if err != nil {
		return err
	}
	if err := bkt.Put(key, raw); err != nil {
		return err
	}
	return nil
}

func dbGet(bkt *bolt.Bucket, key []byte, obj gproto.Message) error {
	raw := bkt.Get(key)
	if raw == nil {
		return fmt.Errorf("not exists")
	}
	if err := gproto.Unmarshal(raw, obj); err != nil {
		return err
	}
	return nil
}

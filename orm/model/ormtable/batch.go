package ormtable

import "github.com/cosmos/cosmos-sdk/orm/model/kvstore"

type batchIndexCommitmentWriter struct {
	kvstore.IndexCommitmentStoreWithHooks
	commitmentWriter *batchStoreWriter
	indexWriter      *batchStoreWriter
}

func newBatchIndexCommitmentWriter(store kvstore.IndexCommitmentStoreWithHooks) *batchIndexCommitmentWriter {
	return &batchIndexCommitmentWriter{
		IndexCommitmentStoreWithHooks: store,
		// optimal array capacities are estimated here:
		commitmentWriter: &batchStoreWriter{
			Reader: store.CommitmentStoreReader(),
			writes: make([]batchWriterEntry, 0, 2),
		},
		indexWriter: &batchStoreWriter{
			Reader: store.IndexStoreReader(),
			writes: make([]batchWriterEntry, 0, 16),
		},
	}
}

func (w *batchIndexCommitmentWriter) CommitmentStore() kvstore.Store {
	return w.commitmentWriter
}

func (w *batchIndexCommitmentWriter) IndexStore() kvstore.Store {
	return w.indexWriter
}

// Write flushes any pending writes.
func (w *batchIndexCommitmentWriter) Write() error {
	err := flushWrites(w.IndexCommitmentStoreWithHooks.CommitmentStore(), w.commitmentWriter.writes)
	if err != nil {
		return err
	}

	err = flushWrites(w.IndexCommitmentStoreWithHooks.IndexStore(), w.indexWriter.writes)
	if err != nil {
		return err
	}

	// clear writes
	w.Close()

	return err
}

func flushWrites(writer kvstore.Writer, writes []batchWriterEntry) error {
	for _, write := range writes {
		if !write.delete {
			err := writer.Set(write.key, write.value)
			if err != nil {
				return err
			}
		} else {
			err := writer.Delete(write.key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Close discards any pending writes and should generally be called using
// a defer statement.
func (w *batchIndexCommitmentWriter) Close() {
	w.commitmentWriter.writes = nil
	w.indexWriter.writes = nil
}

type batchWriterEntry struct {
	key, value []byte
	delete     bool
}

type batchStoreWriter struct {
	kvstore.Reader
	writes []batchWriterEntry
}

func (b *batchStoreWriter) Set(key, value []byte) error {
	b.writes = append(b.writes, batchWriterEntry{key: key, value: value})
	return nil
}

func (b *batchStoreWriter) Delete(key []byte) error {
	b.writes = append(b.writes, batchWriterEntry{key: key, delete: true})
	return nil
}

var _ kvstore.IndexCommitmentStore = &batchIndexCommitmentWriter{}

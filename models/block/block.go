package block

import (
	"errors"
	"time"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/models/activeblock"
	"github.com/mrusme/zeit/runtime"
)

var (
	ErrEndBeforeStart  error = errors.New("End is before start")
	ErrAlreadyRunning  error = errors.New("Tracker is already running")
	ErrNothingToResume error = errors.New("Nothing to resume")
)

type Block struct {
	key            string    `json:"-"`
	OwnerKey       string    `json:"owner_key"`
	ProjectSID     string    `json:"project_sid"`
	TaskSID        string    `json:"task_sid"`
	Note           string    `json:"note"`
	TimestampStart time.Time `json:"start"`
	TimestampEnd   time.Time `json:"end"`
}

func New(ownerKey string) (*Block, error) {
	b := new(Block)
	b.key = database.NewKey(b)
	b.OwnerKey = ownerKey
	return b, nil
}

func (b *Block) SetKey(k string) {
	b.key = k
}

func (b *Block) GetKey() string {
	return b.key
}

func (b *Block) FromProcessedArgs(pa *argsparser.ParsedArgs) error {
	if pa.WasProcessed() == false {
		return errors.New("Unprocessed ParsedArgs")
	}

	b.ProjectSID = pa.ProjectSID
	b.TaskSID = pa.TaskSID
	b.Note = pa.Note
	b.TimestampStart = pa.GetTimestampStart()
	b.TimestampEnd = pa.GetTimestampEnd()

	return nil
}

func Start(rt *runtime.Runtime, b *Block) error {
	var err error

	if b.TimestampStart.IsZero() {
		b.TimestampStart = time.Now()
	}

	if b.TimestampEnd.IsZero() == false && b.TimestampEnd.Before(b.TimestampStart) {
		return ErrEndBeforeStart
	}

	// We call End first to End any currently active Block
	eb := new(Block)
	eb.TimestampEnd = b.TimestampStart.Add(-1 * time.Second)
	if err = End(rt, eb); err != nil {
		return err
	}

	// TODO: This should be one transaction
	// {
	ab, _ := activeblock.Get(rt)
	ab.SetActiveBlockKey(b.GetKey())
	if err = activeblock.Set(rt, ab); err != nil {
		// We couldn't upsert the ActiveBlock, so we fail fully
		return err
	}

	if err = rt.Database.UpsertRowAsStruct(b); err != nil {
		// We couldn't upsert the Block, so we fail fully
		return err
	}
	// }

	return nil
}

func Switch(rt *runtime.Runtime, b *Block) error {
	return Start(rt, b)
}

func Resume(rt *runtime.Runtime, b *Block) error {
	var ab *activeblock.ActiveBlock
	var err error

	if ab, err = activeblock.Get(rt); err != nil {
		return err
	}

	if ab.HasActiveBlockKey() == true {
		return ErrAlreadyRunning
	}

	if ab.HasPreviousBlockKey() == false {
		return ErrNothingToResume
	}

	pbk := ab.GetPreviousBlockKey()

	pb := new(Block)
	err = rt.Database.GetRowAsStruct(pbk, pb)
	if err != nil && rt.Database.ErrIsKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return err
	}

	if err != nil && rt.Database.ErrIsKeyNotFound(err) == true {
		// The previous block apparently doesn't exist anymore, hence we cannot
		// resume it:
		return ErrNothingToResume
	}

	b.ProjectSID = pb.ProjectSID
	b.TaskSID = pb.TaskSID

	return Start(rt, b)
}

func End(rt *runtime.Runtime, b *Block) error {
	var ab *activeblock.ActiveBlock
	var err error

	if ab, err = activeblock.Get(rt); err != nil {
		return err
	}

	abk := ab.GetActiveBlockKey()
	if abk == "" {
		// No active block, we're done
		return nil
	}

	eb := new(Block)
	err = rt.Database.GetRowAsStruct(abk, eb)
	if err != nil && rt.Database.ErrIsKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return err
	}

	if err != nil && rt.Database.ErrIsKeyNotFound(err) == true {
		// We encountered a situation in which there is an ActiveBlock for a
		// Block that doesn't seem to exist anymore. Let's clear the ActiveBlock.
		ab.ClearActiveBlockKey()
		if err = activeblock.Set(rt, ab); err != nil {
			// Okay, well, that sucks
			return err
		}
		// We have cleared the ActiveBlock, hence we're done
		return nil
	}

	// We have found our Block, let's end it. However, we will only end the
	// Block if it hasn't been ended already.
	if eb.TimestampEnd.IsZero() == true {
		if b.TimestampEnd.IsZero() {
			b.TimestampEnd = time.Now()
		}

		if b.TimestampEnd.Before(eb.TimestampStart) {
			return ErrEndBeforeStart
		}
		eb.TimestampEnd = b.TimestampEnd

		// If the Block contains other updates let's apply them as well
		if b.Note != "" {
			eb.Note = b.Note
		}
		// TODO: Do we want to allow users to adjust the TimestampStart when ending
		// a block? It could be handy, it might however overcomplicate things.
		// Adjustments could instead be made from a dedicated `edit` command.

		if err = rt.Database.UpsertRowAsStruct(eb); err != nil {
			// We couldn't persist the change, so we're keeping the ActiveBlock as it is
			return err
		}
	}

	// We have persisted the change (or the Block was ended already -- weird!),
	// so let's clear the ActiveBlock
	ab.ClearActiveBlockKey()
	if err = activeblock.Set(rt, ab); err != nil {
		return err
	}

	return nil
}

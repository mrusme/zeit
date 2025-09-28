package block

import (
	"errors"
	"time"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/val"
	"github.com/mrusme/zeit/models/activeblock"
)

type Block struct {
	key            string    `json:"-"`
	OwnerKey       string    `json:"owner_key"`
	ProjectSID     string    `json:"project_sid" validate:"required,sid,max=32"`
	TaskSID        string    `json:"task_sid" validate:"required,sid,max=32"`
	Note           string    `json:"note" validate:"max=65536"`
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

func List(db *database.Database) (map[string]*Block, error) {
	var err error

	var rows map[string]*Block = make(map[string]*Block)
	if err = database.GetPrefixedRowsAsStruct(
		db,
		database.PrefixForModel(&Block{}),
		rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}

func ListForProjectSID(
	db *database.Database,
	projectSID string,
) (map[string]*Block, error) {
	var lst map[string]*Block = make(map[string]*Block)

	rows, err := List(db)
	if err != nil {
		return nil, err
	}

	for key := range rows {
		if rows[key].ProjectSID == projectSID {
			lst[key] = rows[key]
		}
	}

	return lst, nil
}

func ListForProjectTaskSID(
	db *database.Database,
	projectSID string,
	taskSID string,
) (map[string]*Block, error) {
	var lst map[string]*Block = make(map[string]*Block)

	rows, err := List(db)
	if err != nil {
		return nil, err
	}

	for key := range rows {
		if rows[key].ProjectSID == projectSID &&
			rows[key].TaskSID == taskSID {
			lst[key] = rows[key]
		}
	}

	return lst, nil
}

func Get(db *database.Database, key string) (*Block, error) {
	var err error

	b := new(Block)
	err = db.GetRowAsStruct(key, b)
	if err != nil && db.IsErrKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return nil, err
	} else if err != nil && db.IsErrKeyNotFound(err) == true {
		return nil, errs.ErrKeyNotFound
	}

	return b, nil
}

func Set(db *database.Database, b *Block) error {
	var err error

	if err = val.Validate(*b); err != nil {
		return err
	}

	if err := db.UpsertRowAsStruct(b); err != nil {
		return err
	}

	return nil
}

func GetActive(db *database.Database) (
	bool,
	*activeblock.ActiveBlock,
	*Block,
	error,
) {
	var ab *activeblock.ActiveBlock
	var err error

	if ab, err = activeblock.Get(db); err != nil {
		return false, nil, nil, err
	}

	if ab.HasActiveBlockKey() == false {
		// No active block, we're done
		return false, ab, nil, nil
	}

	var eb *Block
	eb, err = Get(db, ab.GetActiveBlockKey())
	if err != nil && err != errs.ErrKeyNotFound {
		// We encountered an error which is not KeyNotFound
		return false, ab, nil, err
	} else if err != nil && err == errs.ErrKeyNotFound {
		// We encountered a situation in which there is an ActiveBlock for a
		// Block that doesn't seem to exist anymore. Let's clear the ActiveBlock.
		ab.ClearActiveBlockKey()
		if err = activeblock.Set(db, ab); err != nil {
			// Okay, well, that sucks
			return false, ab, nil, err
		}
		// We have cleared the ActiveBlock, hence we're done
		return false, ab, nil, nil
	}

	return true, ab, eb, nil
}

func Start(db *database.Database, b *Block) (*Block, error) {
	var err error

	if b.TimestampStart.IsZero() {
		b.TimestampStart = time.Now()
	}

	if b.TimestampEnd.IsZero() == false && b.TimestampEnd.Before(b.TimestampStart) {
		return nil, errs.ErrEndBeforeStart
	}

	// Even though `Set()` will validate b, we have to do it manually before we
	// call `End(db, eb)` (see below), as otherwise the currently tracking block
	// might get stopped without the new block being started (created) due to
	// a validation issue.
	if err = val.Validate(*b); err != nil {
		return nil, err
	}

	// We call End first to End any currently active Block
	eb := new(Block)
	eb.TimestampEnd = b.TimestampStart.Add(-1 * time.Second)
	err = End(db, eb)
	if err != nil && err != errs.ErrNothingToEnd {
		return nil, err
	}

	// TODO: This should be one transaction
	// {
	ab, _ := activeblock.Get(db)
	ab.SetActiveBlockKey(b.GetKey())
	if err = activeblock.Set(db, ab); err != nil {
		// We couldn't upsert the ActiveBlock, so we fail fully
		return nil, err
	}

	if err = Set(db, b); err != nil {
		// We couldn't upsert the Block, so we fail fully
		return nil, err
	}
	// }

	return b, nil
}

func Switch(db *database.Database, b *Block) (*Block, error) {
	return Start(db, b)
}

func Resume(db *database.Database, b *Block) (*Block, error) {
	var ab *activeblock.ActiveBlock
	var err error

	if ab, err = activeblock.Get(db); err != nil {
		return nil, err
	}

	if ab.HasActiveBlockKey() == true {
		return nil, errs.ErrAlreadyRunning
	}

	if ab.HasPreviousBlockKey() == false {
		return nil, errs.ErrNothingToResume
	}

	pbk := ab.GetPreviousBlockKey()

	var pb *Block
	pb, err = Get(db, pbk)
	if err != nil && err != errs.ErrKeyNotFound {
		// We encountered an error which is not KeyNotFound
		return nil, err
	} else if err != nil && err == errs.ErrKeyNotFound {
		// The previous block apparently doesn't exist anymore, hence we cannot
		// resume it:
		return nil, errs.ErrNothingToResume
	}

	b.ProjectSID = pb.ProjectSID
	b.TaskSID = pb.TaskSID

	return Start(db, b)
}

func End(db *database.Database, b *Block) error {
	var err error

	found, ab, eb, err := GetActive(db)
	if err != nil {
		return err
	}

	if err == nil && found == false {
		return errs.ErrNothingToEnd
	}

	// We have found our Block, let's end it. However, we will only end the
	// Block if it hasn't been ended already.
	if eb.TimestampEnd.IsZero() == true {
		if b.TimestampEnd.IsZero() {
			b.TimestampEnd = time.Now()
		}

		if b.TimestampEnd.Before(eb.TimestampStart) {
			return errs.ErrEndBeforeStart
		}
		eb.TimestampEnd = b.TimestampEnd

		// If the Block contains other updates let's apply them as well
		if b.Note != "" {
			eb.Note = b.Note
		}
		// TODO: Do we want to allow users to adjust the TimestampStart when ending
		// a block? It could be handy, it might however overcomplicate things.
		// Adjustments could instead be made from a dedicated `edit` command.

		if err = Set(db, eb); err != nil {
			// We couldn't persist the change, so we're keeping the ActiveBlock as it is
			return err
		}
	}

	// We have persisted the change (or the Block was ended already -- weird!),
	// so let's clear the ActiveBlock
	ab.ClearActiveBlockKey()
	if err = activeblock.Set(db, ab); err != nil {
		return err
	}

	return nil
}

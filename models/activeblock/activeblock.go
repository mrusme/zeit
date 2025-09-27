package activeblock

import (
	"github.com/mrusme/zeit/database"
)

const KEY string = "activeblock"

type ActiveBlock struct {
	key              string `json:"-"`
	ActiveBlockKey   string `json:"active_block_key"`
	PreviousBlockKey string `json:"previous_block_key"`
}

func New() (*ActiveBlock, error) {
	ab := new(ActiveBlock)
	ab.key = KEY
	ab.ActiveBlockKey = ""
	ab.PreviousBlockKey = ""
	return ab, nil
}

func (ab *ActiveBlock) SetKey(k string) {
	// We won't allow this as this entity is a single static entry
	return
}

func (ab *ActiveBlock) GetKey() string {
	return ab.key
}

func (ab *ActiveBlock) SetActiveBlockKey(k string) {
	if ab.ActiveBlockKey != "" {
		ab.PreviousBlockKey = ab.ActiveBlockKey
	}
	ab.ActiveBlockKey = k
	return
}

func (ab *ActiveBlock) ClearActiveBlockKey() {
	ab.SetActiveBlockKey("")
	return
}

func (ab *ActiveBlock) GetActiveBlockKey() string {
	return ab.ActiveBlockKey
}

func (ab *ActiveBlock) HasActiveBlockKey() bool {
	return ab.GetActiveBlockKey() != ""
}

func (ab *ActiveBlock) GetPreviousBlockKey() string {
	return ab.PreviousBlockKey
}

func (ab *ActiveBlock) HasPreviousBlockKey() bool {
	return ab.GetPreviousBlockKey() != ""
}

func Get(db *database.Database) (*ActiveBlock, error) {
	var err error

	ab, _ := New()
	err = db.GetRowAsStruct(ab.GetKey(), ab)
	if err != nil && db.IsErrKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return nil, err
	}

	// First time users won't have an ActiveBlock, hence we will retrieve an error
	// that is of type KeyNotFound. In that case we would return a New()
	// ActiveBlock, which just so happens to be in `ab` anyway, hence we don't
	// need to handle that case.
	return ab, nil
}

func Set(db *database.Database, ab *ActiveBlock) error {
	if err := db.UpsertRowAsStruct(ab); err != nil {
		return err
	}

	return nil
}


import "time"

type Block struct {
	key      string    `json:"-"`
	Project  string    `json:"project"`
	Task     string    `json:"task"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Notes    string    `json:"notes"`
	OwnerKey string    `json:"owner_key"`
}

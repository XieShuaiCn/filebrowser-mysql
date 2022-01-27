package bolt

import (
	"github.com/asdine/storm"
)

// InitDB 初始化链接
func InitDB(path string) (*storm.DB, error) {
	return storm.Open(path)
}

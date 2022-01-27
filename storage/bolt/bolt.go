package bolt

import (
	"log"

	"github.com/asdine/storm"
)

// InitDB 初始化链接
func InitDB(path string) (*storm.DB, error) {
	log.Println("Bolt url: "+ path)
	return storm.Open(path)
}

package dbresolver

import (
	"gorm.io/gorm"
	"math/rand"
)

type WeightPolicy struct {
}

func (WeightPolicy) Resolve(connPools []gorm.ConnPool) gorm.ConnPool {
	if len(connPools) >= 1 {
		return connPools[0]
	} else {
		return connPools[rand.Intn(len(connPools))]
	}
}


// TODO PROBABLY DELETE THIS
package body

import "sync"

type bodyMap struct {
	bodies map[int]Body
	lock   sync.RWMutex
}

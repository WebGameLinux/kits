package Libs

import (
		. "github.com/uniplaces/carbon"
		"time"
)

type CarbonImpl struct {
		Instance *Carbon
}

func CarbonOf() *CarbonImpl {
		var (
				t   = time.Now()
				obj = new(CarbonImpl)
		)
		obj.Instance = NewCarbon(t)
		return obj
}

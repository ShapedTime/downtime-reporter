package transformers

import (
	"downtime-reporter/core"
	"fmt"
	"strconv"
	"time"
)

type DateTransformer core.Result

func (r DateTransformer) Transform() core.Result {
	newR := core.Result{
		Values:       make(map[string][]string),
		KeysSlice:    make([]string, len(r.KeysSlice)),
		Transformers: append(r.Transformers, "DateTransformer"),
	}
	for i, key := range r.KeysSlice {
		keyInt, err := strconv.Atoi(key)
		if err != nil {
			panic(fmt.Sprintf("cannot convert key %s: %v", key, err))
		}

		t := time.Unix(int64(keyInt), 0).Format(core.DateFormat)

		newR.KeysSlice[i] = t
		newR.Values[t] = r.Values[key]
	}

	return newR
}

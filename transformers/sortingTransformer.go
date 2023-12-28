package transformers

import (
	"downtime-reporter/core"
	"fmt"
	"sort"
	"strconv"
)

type SortingTransformer core.Result

func (r SortingTransformer) Transform() core.Result {
	keys := r.KeysSlice
	sort.Slice(keys, func(i, j int) bool {
		iK, err := strconv.Atoi(keys[i])
		if err != nil {
			panic(fmt.Sprintf("cannot convert key %s: %v", keys[i], err))
		}
		jK, err := strconv.Atoi(keys[j])
		if err != nil {
			panic(fmt.Sprintf("cannot convert key %s: %v", keys[i], err))
		}
		return iK < jK
	})

	return core.Result{
		Values:       r.Values,
		KeysSlice:    keys,
		Transformers: append(r.Transformers, "SortingTransformer"),
	}
}

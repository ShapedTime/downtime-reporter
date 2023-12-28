package transformers

import (
	"downtime-reporter/core"
	"fmt"
	"strconv"
	"time"
)

type ConciseTransformer core.Result

func (r ConciseTransformer) Transform() core.Result {
	var err error

	newR := core.Result{
		Values:       make(map[string][]string),
		KeysSlice:    make([]string, 0),
		Transformers: append(r.Transformers, "ConciseTransformer"),
	}

	isUnix := true

	for _, transformerName := range r.Transformers {
		if transformerName == "DateTransformer" {
			isUnix = false
		}
	}

	sumMinutes := 0.0
	isDown := false
	for i, key := range r.KeysSlice {
		if r.Values[key][0] == "0" && !isDown {
			newR.KeysSlice = append(newR.KeysSlice, key)
			isDown = true
		}

		if (r.Values[key][0] == "1" ||
			i >= (len(r.KeysSlice)-1)) &&
			isDown {
			startTime := newR.KeysSlice[len(newR.KeysSlice)-1]

			var startT time.Time
			var endT time.Time

			if isUnix {
				start, err := strconv.Atoi(startTime)
				if err != nil {
					panic(fmt.Sprintf("cannot convert key %s: %v", key, err))
				}
				end, err := strconv.Atoi(key)
				if err != nil {
					panic(fmt.Sprintf("cannot convert key %s: %v", key, err))
				}

				startT = time.Unix(int64(start), 0)
				endT = time.Unix(int64(end), 0)

			} else {
				startT, err = time.Parse(core.DateFormat, startTime)
				if err != nil {
					panic(fmt.Sprintf("cannot parse date %s: %v", key, err))
				}

				endT, err = time.Parse(core.DateFormat, key)
				if err != nil {
					panic(fmt.Sprintf("cannot parse date %s: %v", key, err))
				}

			}

			timeDiffT := endT.Sub(startT)
			timeDiff := fmt.Sprintf("%.f mins", timeDiffT.Minutes())

			newR.Values[startTime] = []string{key, timeDiff}
			isDown = false
			sumMinutes += timeDiffT.Minutes()
		}
	}

	newR.KeysSlice = append(newR.KeysSlice, "Total")

	// TODO: Add given start time and end time here
	newR.Values["Total"] = []string{fmt.Sprintf("%.f mins", sumMinutes)}

	return newR
}

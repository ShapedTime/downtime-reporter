package writer

import (
	"downtime-reporter/core"
	"fmt"
	"os"
)

type CSV struct {
	fileName string
}

func NewCSVWriter(fileName string) *CSV {
	return &CSV{fileName: fileName}
}

func (w *CSV) Write(r core.Result) error {
	f, err := os.Create(w.fileName)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("cannot open file '%s' for writing: %v", w.fileName, err)
	}

	for _, key := range r.KeysSlice {
		_, err := f.WriteString(fmt.Sprintf("%s; ", key))
		if err != nil {
			return fmt.Errorf("cannot write to file '%s': %v", w.fileName, err)
		}

		for i, value := range r.Values[key] {
			// last value
			if i == len(r.Values[key])-1 {
				_, err := f.WriteString(fmt.Sprintf("%s\n", value))
				if err != nil {
					return fmt.Errorf("cannot write to file '%s': %v", w.fileName, err)
				}
				continue
			}

			_, err := f.WriteString(fmt.Sprintf("%s; ", value))
			if err != nil {
				return fmt.Errorf("cannot write to file '%s': %v", w.fileName, err)
			}
		}
	}

	return nil
}

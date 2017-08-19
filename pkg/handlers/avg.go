package handlers

import "fmt"

type Avg struct {
	Avg float32 `json:"avg"`
}

func (a *Avg) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("{\"avg\": %.5f}", a.Avg)
	return []byte(str), nil
}

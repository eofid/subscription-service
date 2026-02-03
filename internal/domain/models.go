package domain

import (
	"encoding/json"
	"strings"
	"time"
)

type CustomDate time.Time

func (c *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return err
	}
	*c = CustomDate(t)
	return nil
}

func (c CustomDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(c).Format("01-2006"))
}

func (c CustomDate) Time() time.Time {
	return time.Time(c)
}

type Subscription struct {
	ID          int         `json:"id"`
	ServiceName string      `json:"service_name"`
	Price       int         `json:"price"`
	UserID      string      `json:"user_id"`
	StartDate   CustomDate  `json:"start_date"`
	EndDate     *CustomDate `json:"end_date,omitempty"`
}

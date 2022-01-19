package timeformat

import "time"

type Time time.Time

const (
	timeFormat = "2006-01-02 15:04:05"
	dateFormat = "2006-01-02"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	format := timeFormat
	if len(data) <= 12 {
		format = dateFormat
	}
	tp, err := time.ParseInLocation(`"`+format+`"`, string(data), time.Local)
	if err != nil {
		*t = Time{}
	} else {
		*t = Time(tp)
	}
	return nil
}

func (t Time) Time() time.Time {
	return time.Time(t)
}

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	if t.IsZero() == false {
		b = time.Time(t).AppendFormat(b, timeFormat)
	}
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t Time) After(t2 Time) bool {
	return time.Time(t).After(time.Time(t2))
}

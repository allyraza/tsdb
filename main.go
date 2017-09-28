package main

import (
	"fmt"
	"time"
	"math"
)

type Bucket struct {
	Time time.Time `json:"time"`
	Value float64 `json:"value"`
	Count int `json:"count"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Series struct {
	Duration time.Duration
	Resolution time.Duration
	buckets map[int64]*Bucket
}

func NewSeries(duration time.Duration, resolution time.Duration) *Series {
	return &Series{
		Duration: duration,
		Resolution: resolution,
		buckets: make(map[int64]*Bucket, 0),
	}
}

func (s *Series) floor(t time.Time) time.Time {
	return t.Truncate(s.Resolution)
}

func (s *Series) index(t time.Time) int64 {
	return int64(math.Mod(float64(s.floor(t).Unix()), float64(s.Duration.Seconds())))
}

func (s *Series) get(t time.Time) *Bucket {
	floor := s.floor(t)
	idx := s.index(t)

	bucket := s.buckets[idx]
	if bucket == nil || bucket.Time != floor {
		b := &Bucket{floor, 0, 0, 0, 0}
		s.buckets[idx] = b
		return b
	}

	return bucket
}

func (s *Series) Insert(t time.Time, value float64) {
	b := s.get(t)
	b.Value += value
	b.Count++

	if b.Count == 1 {
		b.Min = value
		b.Max = value

		return
	}

	if value < b.Min {
		b.Min = value
	}

	if value > b.Max {
		b.Max = value
	}
}

func (s *Series) Range(start time.Time, end time.Time) []*Bucket {
	var buckets []*Bucket
	startFloor := s.floor(start)
	endFloor := s.floor(end)

	now := time.Now()
	firstPossibleFloor := s.floor(now.Add(-1 * s.Duration))

	for x := startFloor; x.Before(endFloor) || x.Equal(endFloor); x = x.Add(s.Resolution) {
		if x.Before(firstPossibleFloor) || x.After(now) {
			continue
		}

		bucket := s.get(x)
		buckets = append(buckets, bucket)
	}

	return buckets
}

func (s *Series) FromDuration(d time.Duration) []*Bucket {
	now := time.Now()
	start := now.Add(-1 * d).Add(s.Resolution)
	return s.Range(start, now)
}

func main() {
	s := NewSeries(4 * time.Hour, 5 * time.Second)
	s.Insert(time.Now(), 45)
	buckets := s.FromDuration(1 * time.Hour)

	for _, b := range buckets {
		fmt.Println(b)
	}
}

package server

type Limits struct {
	MaxLists       int  // 0 = unlimited
	MaxSubscribers int  // total
	MaxSendsMonth  int  // 0 = unlimited
	OpenTracking   bool
	ClickTracking  bool
	RetentionDays  int
}

// DefaultLimits returns fully-unlocked limits for the standalone edition.
func DefaultLimits() Limits {
	return Limits{
	MaxLists:       0,
	MaxSubscribers: 0,
	MaxSendsMonth:  0,
	OpenTracking:   true,
	ClickTracking:  true,
	RetentionDays:  365,
}
}

// LimitReached returns true if the current count meets or exceeds the limit.
// A limit of 0 is treated as unlimited.
func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}

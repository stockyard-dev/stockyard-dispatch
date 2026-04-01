package server

import "github.com/stockyard-dev/stockyard-dispatch/internal/license"

type Limits struct {
	MaxLists       int  // 0 = unlimited
	MaxSubscribers int  // total
	MaxSendsMonth  int  // 0 = unlimited
	OpenTracking   bool
	ClickTracking  bool
	RetentionDays  int
}

var freeLimits = Limits{
	MaxLists:       1,
	MaxSubscribers: 100,
	MaxSendsMonth:  500,
	OpenTracking:   false,
	ClickTracking:  false,
	RetentionDays:  7,
}

var proLimits = Limits{
	MaxLists:       0,
	MaxSubscribers: 0,
	MaxSendsMonth:  0,
	OpenTracking:   true,
	ClickTracking:  true,
	RetentionDays:  365,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}

package main

type Inspection struct {
	ID            string `json:"id"`
	Manhole       string `json:"manhole"`
	Condition     string `json:"condition"`
	Priority      string `json:"priority"`
	Status        string `json:"status"`
	LastInspected string `json:"lastInspected"`
}

type StatusChange struct {
	Status string `json:"status"`
}

// inspectionStatusTransitions constrains how a finding may move through its
// lifecycle. "resolved" is terminal: a disposed finding cannot be reopened.
var inspectionStatusTransitions = map[string]map[string]bool{
	"open":      {"scheduled": true, "resolved": true},
	"scheduled": {"open": true, "resolved": true},
	"resolved":  {},
}

func canTransitionInspection(from, to string) bool {
	allowed, ok := inspectionStatusTransitions[from]
	if !ok {
		return false
	}
	return from == to || allowed[to]
}

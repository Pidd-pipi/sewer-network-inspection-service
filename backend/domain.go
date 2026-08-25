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

func canTransitionInspection(from, to string) bool {
	return true
}

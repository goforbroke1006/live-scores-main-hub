package model

type Scores struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

type Updates struct {
	Home   string `json:"home"`
	Away   string `json:"away"`
	Scores Scores `json:"scores"`
}

package openai

import "strings"

type ModerationRequest struct {
	Input string `json:"input"`
}

type ModerationResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Results []Result  `json:"results"`
	Error   HTTPError `json:"error"`
}

type Result struct {
	Categories struct {
		Hate                  bool `json:"hate"`
		HateThreatening       bool `json:"hate/threatening"`
		Harassment            bool `json:"harassment"`
		HarassmentThreatening bool `json:"harassment/threatening"`
		SelfHarm              bool `json:"self-harm"`
		SelfHarmIntent        bool `json:"self-harm/intent"`
		SelfHarmInstructions  bool `json:"self-harm/intstructions"`
		Sexual                bool `json:"sexual"`
		SexualMinors          bool `json:"sexual/minors"`
		Violence              bool `json:"violence"`
		ViolenceGraphic       bool `json:"violence/graphic"`
	} `json:"categories"`
	CategoryScores struct {
		Hate                  float64 `json:"hate"`
		HateThreatening       float64 `json:"hate/threatening"`
		Harassment            float64 `json:"harassment"`
		HarassmentThreatening float64 `json:"harassment/threatening"`
		SelfHarm              float64 `json:"self-harm"`
		SelfHarmIntent        float64 `json:"self-harm/intent"`
		SelfHarmInstructions  float64 `json:"self-harm/intstructions"`
		Sexual                float64 `json:"sexual"`
		SexualMinors          float64 `json:"sexual/minors"`
		Violence              float64 `json:"violence"`
		ViolenceGraphic       float64 `json:"violence/graphic"`
	} `json:"category_scores"`
	Flagged bool `json:"flagged"`
}

func (mr *ModerationResponse) IsFlagged() bool {
	for _, res := range mr.Results {
		if res.Flagged {
			return true
		}
	}
	return false
}

func (mr *ModerationResponse) FlaggedReason() string {
	var reasons []string
	for _, res := range mr.Results {
		if res.Flagged {
			if res.Categories.Hate {
				reasons = append(reasons, "Hate")
			}
			if res.Categories.HateThreatening {
				reasons = append(reasons, "Hate/Threatening")
			}
			if res.Categories.Harassment {
				reasons = append(reasons, "Harassment")
			}
			if res.Categories.HarassmentThreatening {
				reasons = append(reasons, "Harassment/Threatening")
			}
			if res.Categories.SelfHarm {
				reasons = append(reasons, "SelfHarm")
			}
			if res.Categories.SelfHarmIntent {
				reasons = append(reasons, "SelfHarm/Intent")
			}
			if res.Categories.SelfHarmInstructions {
				reasons = append(reasons, "SelfHarm/Instructions")
			}
			if res.Categories.Sexual {
				reasons = append(reasons, "Sexual")
			}
			if res.Categories.SexualMinors {
				reasons = append(reasons, "Sexual/Minors")
			}
			if res.Categories.Violence {
				reasons = append(reasons, "Violence")
			}
			if res.Categories.ViolenceGraphic {
				reasons = append(reasons, "Violence/Graphic")
			}
		}
	}

	// @todo filter duplicates
	r := strings.Join(reasons, ",")
	if len(r) == 0 {
		// It can occur for example if OpenAI has added new categories that are not yet supported
		r = "Other"
	}
	return r
}

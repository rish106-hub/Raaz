package main

import "testing"

func TestModerateCleanMessage(t *testing.T) {
	result := Moderate("I love reading books on weekends")
	if result.Flagged {
		t.Errorf("clean message flagged as %q", result.Category)
	}
}

func TestModerateProfanity(t *testing.T) {
	cases := []string{"what the fuck", "that's bullshit", "you're such a bitch"}
	for _, msg := range cases {
		r := Moderate(msg)
		if !r.Flagged || r.Category != CategoryProfanity {
			t.Errorf("%q: expected profanity, got flagged=%v category=%q", msg, r.Flagged, r.Category)
		}
	}
}

func TestModerateHarassment(t *testing.T) {
	cases := []string{"kys", "kill yourself", "you should die", "go die"}
	for _, msg := range cases {
		r := Moderate(msg)
		if !r.Flagged || r.Category != CategoryHarassment {
			t.Errorf("%q: expected harassment, got flagged=%v category=%q", msg, r.Flagged, r.Category)
		}
	}
}

func TestModerateExplicit(t *testing.T) {
	cases := []string{"send me nudes", "send nudes", "porn"}
	for _, msg := range cases {
		r := Moderate(msg)
		if !r.Flagged || r.Category != CategoryExplicit {
			t.Errorf("%q: expected explicit, got flagged=%v category=%q", msg, r.Flagged, r.Category)
		}
	}
}

func TestModerateCrisis(t *testing.T) {
	cases := []string{
		"i want to kill myself",
		"thinking about suicide",
		"i want to die",
		"suicidal thoughts",
		"self-harm every day",
		"no reason to live anymore",
	}
	for _, msg := range cases {
		r := Moderate(msg)
		if !r.Flagged || !r.IsCrisis || r.Category != CategoryCrisis {
			t.Errorf("%q: expected crisis, got flagged=%v isCrisis=%v category=%q",
				msg, r.Flagged, r.IsCrisis, r.Category)
		}
	}
}

func TestModerateCrisisPriorityOverHarassment(t *testing.T) {
	// "kill myself" matches both crisis and could theoretically overlap;
	// crisis must win.
	r := Moderate("I want to kill myself and end my pain")
	if !r.IsCrisis || r.Category != CategoryCrisis {
		t.Errorf("crisis pattern should outrank harassment; got category=%q isCrisis=%v",
			r.Category, r.IsCrisis)
	}
}

func TestModerateCaseInsensitive(t *testing.T) {
	r := Moderate("KILL YOURSELF")
	if !r.Flagged || r.Category != CategoryHarassment {
		t.Errorf("uppercase harassment not caught: flagged=%v category=%q", r.Flagged, r.Category)
	}
}

func TestModerateDoesNotMutateInput(t *testing.T) {
	msg := "Hello World"
	Moderate(msg)
	if msg != "Hello World" {
		t.Error("Moderate must not modify the input string")
	}
}

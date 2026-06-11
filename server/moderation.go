package main

import (
	"regexp"
	"strings"
)

// ContentCategory classifies why a message was flagged.
type ContentCategory string

const (
	CategoryClean      ContentCategory = ""
	CategoryProfanity  ContentCategory = "profanity"
	CategoryHarassment ContentCategory = "harassment"
	CategoryExplicit   ContentCategory = "explicit"
	CategoryCrisis     ContentCategory = "crisis"
)

// ModerationResult is returned by Moderate for each message.
type ModerationResult struct {
	Category ContentCategory
	Flagged  bool
	IsCrisis bool
}

var (
	profanityRe  []*regexp.Regexp
	harassmentRe []*regexp.Regexp
	explicitRe   []*regexp.Regexp
	crisisRe     []*regexp.Regexp
)

func compileAll(patterns []string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		result[i] = regexp.MustCompile(p)
	}
	return result
}

func init() {
	profanityRe = compileAll([]string{
		`\bf+u+c+k(ing|ed|er|ers|s)?\b`,
		`\bs+h+i+t(s|ty|head|bag)?\b`,
		`\bb+i+t+c+h(es|y)?\b`,
		`\ba+s+s+h+o+l+e+s?\b`,
		`\bc+u+n+t+s?\b`,
	})
	harassmentRe = compileAll([]string{
		`\bkys\b`,
		`\bkill\s+yourself\b`,
		`\bgo\s+(die|hang\s+yourself)\b`,
		`\byou\s+should\s+die\b`,
		`\bi\s+(will|am\s+going\s+to|gonna)\s+(kill|hurt|destroy)\s+you\b`,
		`\bi\s+will\s+find\s+you\b`,
	})
	explicitRe = compileAll([]string{
		`\bsend\s+(me\s+)?(your\s+)?nudes?\b`,
		`\bnudes?\b`,
		`\bporn(ography)?\b`,
		`\bsex(ual)?\s+(pic|photo|video)s?\b`,
	})
	// Crisis patterns take priority — checked first in Moderate.
	crisisRe = compileAll([]string{
		`\bsuicid(e|al|ally)\b`,
		`\bself[- ]?harm\b`,
		`\bwant\s+to\s+(die|end\s+(it|my\s+life))\b`,
		`\b(kill|hurt)\s+myself\b`,
		`\bno\s+reason\s+to\s+live\b`,
		`\bcan'?t\s+(go\s+on|take\s+it\s+anymore)\b`,
		`\bend\s+my\s+(life|pain|suffering)\b`,
		`\bthinking\s+(about|of)\s+(suicide|killing\s+myself|ending\s+(it|my\s+life))\b`,
	})
}

// Moderate checks text for policy violations. Input is normalised to lowercase
// before matching; the original text is not modified. Crisis is checked first
// so life-safety events are never misclassified as ordinary violations.
func Moderate(text string) ModerationResult {
	lower := strings.ToLower(text)

	for _, p := range crisisRe {
		if p.MatchString(lower) {
			return ModerationResult{Category: CategoryCrisis, Flagged: true, IsCrisis: true}
		}
	}
	for _, p := range harassmentRe {
		if p.MatchString(lower) {
			return ModerationResult{Category: CategoryHarassment, Flagged: true}
		}
	}
	for _, p := range explicitRe {
		if p.MatchString(lower) {
			return ModerationResult{Category: CategoryExplicit, Flagged: true}
		}
	}
	for _, p := range profanityRe {
		if p.MatchString(lower) {
			return ModerationResult{Category: CategoryProfanity, Flagged: true}
		}
	}
	return ModerationResult{}
}

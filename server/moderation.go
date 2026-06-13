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

	// maxModerateLen caps input length before regex evaluation to prevent ReDoS
	// on extremely long strings fed through bounded-but-chained quantifiers.
	maxModerateLen = 4096
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
	// L-7: use bounded quantifiers {1,N} instead of + to prevent catastrophic backtracking.
	profanityRe = compileAll([]string{
		`\bf{1,10}u{1,10}c{1,10}k{1,10}(ing|ed|er|ers|s)?\b`,
		`\bs{1,10}h{1,10}i{1,10}t{1,10}(s|ty|head|bag)?\b`,
		`\bb{1,10}u{1,10}l{1,10}l{1,10}s{1,10}h{1,10}i{1,10}t{1,10}s?\b`,
		`\bb{1,10}i{1,10}t{1,10}c{1,10}h{1,10}(es|y)?\b`,
		`\ba{1,10}s{1,10}s{1,10}h{1,10}o{1,10}l{1,10}e{1,10}s?\b`,
		`\bc{1,10}u{1,10}n{1,10}t{1,10}s?\b`,
	})
	harassmentRe = compileAll([]string{
		`\bkys\b`,
		`\bkill\s{1,20}yourself\b`,
		`\bgo\s{1,20}(die|hang\s{1,20}yourself)\b`,
		`\byou\s{1,20}should\s{1,20}die\b`,
		`\bi\s{1,20}(will|am\s{1,20}going\s{1,20}to|gonna)\s{1,20}(kill|hurt|destroy)\s{1,20}you\b`,
		`\bi\s{1,20}will\s{1,20}find\s{1,20}you\b`,
	})
	explicitRe = compileAll([]string{
		`\bsend\s{1,20}(me\s{1,20})?(your\s{1,20})?nudes?\b`,
		`\bnudes?\b`,
		`\bporn(ography)?\b`,
		`\bsex(ual)?\s{1,20}(pic|photo|video)s?\b`,
	})
	crisisRe = compileAll([]string{
		`\bsuicid(e|al|ally)\b`,
		`\bself[- ]?harm\b`,
		`\bwant\s{1,20}to\s{1,20}(die|end\s{1,20}(it|my\s{1,20}life))\b`,
		`\b(kill|hurt)\s{1,20}myself\b`,
		`\bno\s{1,20}reason\s{1,20}to\s{1,20}live\b`,
		`\bcan'?t\s{1,20}(go\s{1,20}on|take\s{1,20}it\s{1,20}anymore)\b`,
		`\bend\s{1,20}my\s{1,20}(life|pain|suffering)\b`,
		`\bthinking\s{1,20}(about|of)\s{1,20}(suicide|killing\s{1,20}myself|ending\s{1,20}(it|my\s{1,20}life))\b`,
	})
}

// Moderate checks text for policy violations. Input is normalised to lowercase
// and capped at maxModerateLen. Crisis is checked first so life-safety events
// are never misclassified as ordinary violations.
func Moderate(text string) ModerationResult {
	if len(text) > maxModerateLen {
		text = text[:maxModerateLen]
	}
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

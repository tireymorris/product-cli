package prompt

import "strings"

const (
	kindMarkerPrefix = "===ralph-prompt-kind:"
	kindMarkerSuffix = "==="
)

const (
	KindClarify                  = "clarify"
	KindPRDGenerate              = "prd-generate"
	KindPRDSelfReview            = "prd-self-review"
	KindPRDCritiqueRevision      = "prd-critique-revision"
	KindPRDClarificationRevision = "prd-clarification-revision"
	KindStoryImplement           = "story-implement"
	KindDiffReview               = "diff-review"
	KindRecovery                 = "recovery"
	KindCleanup                  = "cleanup"
	KindFollowUp                 = "followup"
)

func wrapWithKind(kind, body string) string {
	if kind == "" {
		return body
	}
	return kindMarkerPrefix + kind + kindMarkerSuffix + "\n" + body
}

func Kind(prompt string) string {
	_, after, ok := strings.Cut(prompt, kindMarkerPrefix)
	if !ok {
		return ""
	}
	before, _, ok := strings.Cut(after, kindMarkerSuffix)
	if !ok {
		return ""
	}
	return before
}

func HasKind(prompt, kind string) bool {
	return Kind(prompt) == kind
}

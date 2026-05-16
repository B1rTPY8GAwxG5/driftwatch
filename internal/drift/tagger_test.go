package drift

import (
	"testing"
)

func taggerDriftedResult(service string, kinds ...DriftKind) DriftResult {
	var entries []DriftEntry
	for _, k := range kinds {
		entries = append(entries, DriftEntry{Kind: k, Field: string(k)})
	}
	return DriftResult{Service: service, Entries: entries}
}

func TestNewTagger_NotNil(t *testing.T) {
	tgr := NewTagger()
	if tgr == nil {
		t.Fatal("expected non-nil Tagger")
	}
}

func TestTagger_Tag_StaticTagsAlwaysPresent(t *testing.T) {
	tgr := NewTagger("env:prod", "team:platform")
	result := DriftResult{Service: "svc-a"}
	tags := tgr.Tag(result)
	if len(tags) != 2 {
		t.Fatalf("expected 2 static tags, got %d", len(tags))
	}
}

func TestTagger_Tag_RuleMatchingKind(t *testing.T) {
	tgr := NewTagger()
	tgr.AddRule(TaggerRule{Kind: KindImage, Tags: []string{"drift:image"}})
	result := taggerDriftedResult("svc-b", KindImage)
	tags := tgr.Tag(result)
	if len(tags) != 1 || tags[0] != "drift:image" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagger_Tag_RuleNonMatchingKind_NotApplied(t *testing.T) {
	tgr := NewTagger()
	tgr.AddRule(TaggerRule{Kind: KindImage, Tags: []string{"drift:image"}})
	result := taggerDriftedResult("svc-c", KindReplicas)
	tags := tgr.Tag(result)
	if len(tags) != 0 {
		t.Fatalf("expected no tags, got %v", tags)
	}
}

func TestTagger_Tag_EmptyKindRule_AlwaysApplied(t *testing.T) {
	tgr := NewTagger()
	tgr.AddRule(TaggerRule{Kind: "", Tags: []string{"checked"}})
	result := DriftResult{Service: "svc-d"}
	tags := tgr.Tag(result)
	if len(tags) != 1 || tags[0] != "checked" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagger_Tag_DeduplicatesTags(t *testing.T) {
	tgr := NewTagger("shared")
	tgr.AddRule(TaggerRule{Kind: "", Tags: []string{"shared", "extra"}})
	result := DriftResult{Service: "svc-e"}
	tags := tgr.Tag(result)
	// "shared" should appear only once
	count := 0
	for _, tag := range tags {
		if tag == "shared" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 'shared' once, got %d times in %v", count, tags)
	}
}

func TestTagger_AddRule_EmptyTagsIgnored(t *testing.T) {
	tgr := NewTagger()
	tgr.AddRule(TaggerRule{Kind: KindImage, Tags: nil})
	result := taggerDriftedResult("svc-f", KindImage)
	tags := tgr.Tag(result)
	if len(tags) != 0 {
		t.Fatalf("expected no tags from ignored rule, got %v", tags)
	}
}

func TestTagger_TagAll_ReturnsMapByService(t *testing.T) {
	tgr := NewTagger("global")
	results := []DriftResult{
		{Service: "alpha"},
		{Service: "beta"},
	}
	m := tgr.TagAll(results)
	if len(m) != 2 {
		t.Fatalf("expected map length 2, got %d", len(m))
	}
	for _, svc := range []string{"alpha", "beta"} {
		if len(m[svc]) != 1 || m[svc][0] != "global" {
			t.Fatalf("unexpected tags for %s: %v", svc, m[svc])
		}
	}
}

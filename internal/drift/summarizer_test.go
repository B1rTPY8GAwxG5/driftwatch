package drift

import (
	"bytes"
	"strings"
	"testing"
)

func makeSummarizerResult(service string, drifted bool) DriftResult {
	r := DriftResult{Service: service}
	if drifted {
		r.Entries = []DriftEntry{
			{Field: "image", Kind: KindImage, Declared: "v1", Observed: "v2"},
		}
	}
	return r
}

func TestNewSummarizer_DefaultPeriod(t *testing.T) {
	s := NewSummarizer("")
	if s.period != SummaryPeriodDaily {
		t.Errorf("expected daily, got %s", s.period)
	}
}

func TestNewSummarizer_CustomPeriod(t *testing.T) {
	s := NewSummarizer(SummaryPeriodWeekly)
	if s.period != SummaryPeriodWeekly {
		t.Errorf("expected weekly, got %s", s.period)
	}
}

func TestSummarizer_Record_EmptyServiceIgnored(t *testing.T) {
	s := NewSummarizer(SummaryPeriodDaily)
	s.Record(DriftResult{})
	ds := s.Build()
	if len(ds.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(ds.Services))
	}
}

func TestSummarizer_Record_CleanResult(t *testing.T) {
	s := NewSummarizer(SummaryPeriodDaily)
	s.Record(makeSummarizerResult("svc-a", false))
	ds := s.Build()
	if len(ds.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(ds.Services))
	}
	if ds.Services[0].TotalChecks != 1 {
		t.Errorf("expected TotalChecks=1, got %d", ds.Services[0].TotalChecks)
	}
	if ds.Services[0].DriftedChecks != 0 {
		t.Errorf("expected DriftedChecks=0, got %d", ds.Services[0].DriftedChecks)
	}
}

func TestSummarizer_Record_DriftedResult(t *testing.T) {
	s := NewSummarizer(SummaryPeriodDaily)
	s.Record(makeSummarizerResult("svc-b", true))
	ds := s.Build()
	if ds.Services[0].DriftedChecks != 1 {
		t.Errorf("expected DriftedChecks=1, got %d", ds.Services[0].DriftedChecks)
	}
	if ds.Services[0].Kinds[KindImage] != 1 {
		t.Errorf("expected image kind count=1")
	}
}

func TestSummarizer_Build_SortedAlphabetically(t *testing.T) {
	s := NewSummarizer(SummaryPeriodHourly)
	s.Record(makeSummarizerResult("zebra", false))
	s.Record(makeSummarizerResult("alpha", false))
	ds := s.Build()
	if ds.Services[0].Service != "alpha" {
		t.Errorf("expected alpha first, got %s", ds.Services[0].Service)
	}
}

func TestDriftSummary_DriftRate_NoChecks(t *testing.T) {
	ds := DriftSummary{}
	if ds.DriftRate() != 0 {
		t.Errorf("expected 0 rate for empty summary")
	}
}

func TestDriftSummary_DriftRate_Calculated(t *testing.T) {
	s := NewSummarizer(SummaryPeriodDaily)
	s.Record(makeSummarizerResult("svc", true))
	s.Record(makeSummarizerResult("svc", false))
	ds := s.Build()
	rate := ds.DriftRate()
	if rate != 50.0 {
		t.Errorf("expected 50.0, got %.1f", rate)
	}
}

func TestWriteSummary_ContainsPeriod(t *testing.T) {
	s := NewSummarizer(SummaryPeriodWeekly)
	s.Record(makeSummarizerResult("svc-x", true))
	var buf bytes.Buffer
	WriteSummary(&buf, s.Build())
	if !strings.Contains(buf.String(), "weekly") {
		t.Errorf("expected 'weekly' in output, got: %s", buf.String())
	}
}

func TestWriteSummary_ContainsServiceName(t *testing.T) {
	s := NewSummarizer(SummaryPeriodDaily)
	s.Record(makeSummarizerResult("my-service", false))
	var buf bytes.Buffer
	WriteSummary(&buf, s.Build())
	if !strings.Contains(buf.String(), "my-service") {
		t.Errorf("expected service name in output")
	}
}

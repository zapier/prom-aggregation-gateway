package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func TestNewAggregates(t *testing.T) {
	contentType := expfmt.FmtText

	for id, c := range []struct {
		testName      string
		a, b          string
		want          string
		ignoredLabels []string
	}{
		{"simpleGauge", gaugeInput, gaugeInput, gaugeOutput, []string{}},
		{"in", in1, in2, want, []string{}},
		{"multilabel", multilabel1, multilabel2, multilabelResult, []string{"ignore_label"}},
		{"labelFields", labelFields1, labelFields2, labelFieldResult, []string{}},
		{"reorderedLabels", reorderedLabels1, reorderedLabels2, reorderedLabelsResult, []string{}},
		{"ignoredLabels", ignoredLabels1, ignoredLabels2, ignoredLabelsResult, []string{"ignore_me"}},
	} {
		t.Run(fmt.Sprintf("id-%d %s", id, c.testName), func(t *testing.T) {
			aggs := NewAggregates(time.Microsecond*200, AddIgnoredLabels(c.ignoredLabels...))

			err := aggs.writeAggregate.parseAndMerge(strings.NewReader(c.a), testLabels)
			require.NoError(t, err)

			err = aggs.writeAggregate.parseAndMerge(strings.NewReader(c.b), testLabels)
			require.NoError(t, err)

			time.Sleep(time.Microsecond * 400)

			buf := new(bytes.Buffer)
			aggs.readAggregate.encodeAllMetrics(buf, contentType)

			if have := buf.String(); have != c.want {
				text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(c.want),
					B:        difflib.SplitLines(have),
					FromFile: "have",
					ToFile:   "want",
					Context:  3,
				})
				t.Fatalf("%s: %s", c.testName, text)
			}
		})
	}
}

func TestMetricsCopy(t *testing.T) {
	contentType := expfmt.FmtText

	for id, c := range []struct {
		testName      string
		a, b          string
		want1, want2  string
		timeBatch     time.Duration
		timeWait      time.Duration
		timeSleep     time.Duration
		ignoredLabels []string
	}{
		{"multilabel separate batch", multilabel1, multilabel2, multilabelSingleResult, multilabelResult, time.Microsecond * 100, time.Microsecond * 110, time.Microsecond * 150, []string{"ignore_label"}},
	} {
		t.Run(fmt.Sprintf("id-%d %s", id, c.testName), func(t *testing.T) {
			aggs := NewAggregates(c.timeBatch, AddIgnoredLabels(c.ignoredLabels...))

			err := aggs.writeAggregate.parseAndMerge(strings.NewReader(c.a), testLabels)
			require.NoError(t, err)

			time.Sleep(c.timeWait)

			buf := new(bytes.Buffer)
			aggs.readAggregate.encodeAllMetrics(buf, contentType)

			if have := buf.String(); have != c.want1 {
				text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(c.want1),
					B:        difflib.SplitLines(have),
					FromFile: "have",
					ToFile:   "want",
					Context:  3,
				})
				t.Fatalf("%s: %s", c.testName, text)
			}

			time.Sleep(c.timeSleep)

			err = aggs.writeAggregate.parseAndMerge(strings.NewReader(c.b), testLabels)
			require.NoError(t, err)

			time.Sleep(c.timeSleep)

			buf = new(bytes.Buffer)
			aggs.readAggregate.encodeAllMetrics(buf, contentType)

			if have := buf.String(); have != c.want2 {
				text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(c.want2),
					B:        difflib.SplitLines(have),
					FromFile: "have",
					ToFile:   "want",
					Context:  3,
				})
				t.Fatalf("%s: %s", c.testName, text)
			}
		})
	}
}

func TestGaugesAreBatched(t *testing.T) {
	contentType := expfmt.FmtText

	for id, c := range []struct {
		testName      string
		a, b          string
		want          string
		timeBatch     time.Duration
		timeWait      time.Duration
		timeSleep     time.Duration
		ignoredLabels []string
	}{
		{"simpleGauge", gaugeInput, gaugeInput, gaugeOutput, time.Microsecond * 100, time.Nanosecond * 1, time.Microsecond * 200, []string{}},
		{"gaugeBatched", gaugeInput, gaugeInput, gaugeBatchedOutput, time.Microsecond * 100, time.Microsecond * 300, time.Microsecond * 300, []string{}},
	} {
		t.Run(fmt.Sprintf("id-%d %s", id, c.testName), func(t *testing.T) {
			aggs := NewAggregates(c.timeBatch, AddIgnoredLabels(c.ignoredLabels...))

			err := aggs.writeAggregate.parseAndMerge(strings.NewReader(c.a), testLabels)
			require.NoError(t, err)

			time.Sleep(c.timeWait)

			err = aggs.writeAggregate.parseAndMerge(strings.NewReader(c.b), testLabels)
			require.NoError(t, err)

			time.Sleep(c.timeSleep)

			buf := new(bytes.Buffer)
			aggs.readAggregate.encodeAllMetrics(buf, contentType)

			if have := buf.String(); have != c.want {
				text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(c.want),
					B:        difflib.SplitLines(have),
					FromFile: "have",
					ToFile:   "want",
					Context:  3,
				})
				t.Fatalf("%s: %s", c.testName, text)
			}
		})
	}
}

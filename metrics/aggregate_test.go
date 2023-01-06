package metrics

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

var testLabels = []labelPair{
	{"job", "test"},
}

func TestAggregate(t *testing.T) {
	contentType := expfmt.FmtText

	for _, c := range []struct {
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
		t.Run(c.testName, func(t *testing.T) {
			agg := newAggregate(AddIgnoredLabels(c.ignoredLabels...))

			err := agg.parseAndMerge(strings.NewReader(c.a), testLabels)
			require.NoError(t, err)

			err = agg.parseAndMerge(strings.NewReader(c.b), testLabels)
			require.NoError(t, err)

			buf := new(bytes.Buffer)
			agg.encodeAllMetrics(buf, contentType)

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

	t.Run("duplicateLabels", func(t *testing.T) {
		agg := newAggregate()

		err := agg.parseAndMerge(strings.NewReader(duplicateLabels), testLabels)
		require.Equal(t, err.Error(), duplicateError)
	})
}

var testMetricTable = []struct {
	inputName      string
	input1, input2 string
	ignoredLabels  []string
}{
	{"simpleGauge", gaugeInput, gaugeInput, []string{}},
	{"fullMetrics", in1, in2, []string{}},
	{"multiLabel", multilabel1, multilabel2, []string{}},
	{"multiLabelIgnore", multilabel1, multilabel2, []string{"ignore_label"}},
	{"labelFields", labelFields1, labelFields2, []string{}},
	{"reorderedLabels", reorderedLabels1, reorderedLabels2, []string{}},
	{"ignoredLabels", ignoredLabels1, ignoredLabels2, []string{"ignore_me"}},
}

func BenchmarkAggregate(b *testing.B) {
	a := newAggregate()
	for _, v := range testMetricTable {
		a.options.ignoredLabels = v.ignoredLabels
		b.Run(fmt.Sprintf("metric_type_%s", v.inputName), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if err := a.parseAndMerge(strings.NewReader(v.input1), testLabels); err != nil {
					b.Fatalf("unexpected error %s", err)
				}
				if err := a.parseAndMerge(strings.NewReader(v.input2), testLabels); err != nil {
					b.Fatalf("unexpected error %s", err)
				}
			}
		})
	}
}

func BenchmarkConcurrentAggregate(b *testing.B) {
	a := newAggregate()
	for _, v := range testMetricTable {
		a.options.ignoredLabels = v.ignoredLabels
		b.Run(fmt.Sprintf("metric_type_%s", v.inputName), func(b *testing.B) {
			if err := a.parseAndMerge(strings.NewReader(v.input1), testLabels); err != nil {
				b.Fatalf("unexpected error %s", err)
			}

			for n := 0; n < b.N; n++ {
				g, _ := errgroup.WithContext(context.Background())
				for tN := 0; tN < 10; tN++ {
					g.Go(func() error {
						return a.parseAndMerge(strings.NewReader(v.input2), testLabels)
					})
				}

				if err := g.Wait(); err != nil {
					b.Fatalf("unexpected error %s", err)
				}

			}
		})
	}
}

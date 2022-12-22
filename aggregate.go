package main

import (
	"io"
	"log"
	"sort"
	"strings"
	"sync"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type metricFamily struct {
	*dto.MetricFamily
	lock sync.RWMutex
}

func (mf *metricFamily) Len() int {
	mf.lock.RLock()
	count := len(mf.Metric)
	mf.lock.RUnlock()

	return count
}

type aggregate struct {
	familiesLock sync.RWMutex
	families     map[string]*metricFamily
	options      aggregateOptions
}

type ignoredLabels []string

type aggregateOptions struct {
	ignoredLabels ignoredLabels
}

type AggregateOptionsFunc func(a *aggregate)

func AddIgnoredLabels(ignoredLabels ...string) AggregateOptionsFunc {
	return func(a *aggregate) {
		a.options.ignoredLabels = ignoredLabels
	}
}

func newAggregate(opts ...AggregateOptionsFunc) *aggregate {
	a := &aggregate{
		families: map[string]*metricFamily{},
		options: aggregateOptions{
			ignoredLabels: []string{},
		},
	}

	for _, opt := range opts {
		opt(a)
	}

	a.options.formatOptions()

	return a
}

func (ao *aggregateOptions) formatOptions() {
	ao.formatIgnoredLabels()
}

func (ao *aggregateOptions) formatIgnoredLabels() {
	if ao.ignoredLabels != nil {
		for i, v := range ao.ignoredLabels {
			ao.ignoredLabels[i] = strings.ToLower(v)
		}
	}

	sort.Strings(ao.ignoredLabels)
}

func (a *aggregate) Len() int {
	a.familiesLock.RLock()
	count := len(a.families)
	a.familiesLock.RUnlock()
	return count
}

// setFamilyOrGetExistingFamily either sets a new family or returns an existing family
func (a *aggregate) setFamilyOrGetExistingFamily(familyName string, family *dto.MetricFamily) *metricFamily {
	a.familiesLock.Lock()
	defer a.familiesLock.Unlock()
	existingFamily, ok := a.families[familyName]
	if !ok {
		a.families[familyName] = &metricFamily{MetricFamily: family}
		return nil
	}
	return existingFamily
}

func (a *aggregate) saveFamily(familyName string, family *dto.MetricFamily) error {
	existingFamily := a.setFamilyOrGetExistingFamily(familyName, family)
	if existingFamily != nil {
		err := existingFamily.mergeFamily(family)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *aggregate) parseAndMerge(r io.Reader, job string) error {
	var parser expfmt.TextParser
	inFamilies, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return err
	}

	for name, family := range inFamilies {
		// Sort labels in case source sends them inconsistently
		for _, m := range family.Metric {
			a.formatLabels(m, job)
		}

		if err := validateFamily(family); err != nil {
			return err
		}

		// family must be sorted for the merge
		sort.Sort(byLabel(family.Metric))

		if err := a.saveFamily(name, family); err != nil {
			return err
		}

		MetricCountByFamily.WithLabelValues(name).Set(float64(len(family.Metric)))

	}

	TotalFamiliesGauge.Set(float64(a.Len()))

	return nil
}

func (a *aggregate) render(enc expfmt.Encoder) {
	a.familiesLock.RLock()
	defer a.familiesLock.RUnlock()

	var metricNames []string
	metricTypeCounts := make(map[string]int)
	metricFamilyCounts := make(map[string]int)
	for name, family := range a.families {
		metricNames = append(metricNames, name)
		var typeName string
		if family.Type == nil {
			typeName = "unknown"
		} else {
			typeName = dto.MetricType_name[int32(*family.Type)]
		}
		metricTypeCounts[typeName]++
		metricFamilyCounts[name] = family.Len()
	}

	sort.Strings(metricNames)

	for _, name := range metricNames {
		if a.encodeMetric(name, enc) {
			return
		}
	}

	MetricCountByType.Reset()
	for typeName, count := range metricTypeCounts {
		MetricCountByType.WithLabelValues(typeName).Set(float64(count))
	}

	MetricCountByFamily.Reset()
	for familyName, count := range metricFamilyCounts {
		MetricCountByFamily.WithLabelValues(familyName).Set(float64(count))
	}
}

func (a *aggregate) encodeMetric(name string, enc expfmt.Encoder) bool {
	a.families[name].lock.RLock()
	defer a.families[name].lock.RUnlock()

	if err := enc.Encode(a.families[name].MetricFamily); err != nil {
		log.Printf("An error has occurred during metrics encoding:\n\n%s\n", err.Error())
		return true
	}
	return false
}

package metrics

const (
	in1 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 42
# HELP counter A counter
# TYPE counter counter
counter 31
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0
histogram_bucket{le="2"} 0
histogram_bucket{le="3"} 3
histogram_bucket{le="4"} 4
histogram_bucket{le="5"} 4
histogram_bucket{le="6"} 4
histogram_bucket{le="7"} 4
histogram_bucket{le="8"} 4
histogram_bucket{le="9"} 4
histogram_bucket{le="10"} 4
histogram_bucket{le="+Inf"} 4
histogram_sum{} 2.5
histogram_count{} 1
`
	in2 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 57
# HELP counter A counter
# TYPE counter counter
counter 29
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0
histogram_bucket{le="2"} 0
histogram_bucket{le="3"} 0
histogram_bucket{le="4"} 4
histogram_bucket{le="5"} 5
histogram_bucket{le="6"} 5
histogram_bucket{le="7"} 5
histogram_bucket{le="8"} 5
histogram_bucket{le="9"} 5
histogram_bucket{le="10"} 5
histogram_bucket{le="+Inf"} 5
histogram_sum 4.5
histogram_count 1
`
	want = `# HELP counter A counter
# TYPE counter counter
counter{job="test"} 60
# HELP gauge A gauge
# TYPE gauge gauge
gauge{job="test"} 99
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{job="test",le="1"} 0
histogram_bucket{job="test",le="2"} 0
histogram_bucket{job="test",le="3"} 3
histogram_bucket{job="test",le="4"} 8
histogram_bucket{job="test",le="5"} 9
histogram_bucket{job="test",le="6"} 9
histogram_bucket{job="test",le="7"} 9
histogram_bucket{job="test",le="8"} 9
histogram_bucket{job="test",le="9"} 9
histogram_bucket{job="test",le="10"} 9
histogram_bucket{job="test",le="+Inf"} 9
histogram_sum{job="test"} 7
histogram_count{job="test"} 2
`

	multilabel1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b", ignore_label="ignore_value"} 1
`
	multilabel2 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b", ignore_label="ignore_value"} 2
`
	multilabelSingleResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b",job="test"} 1
`
	multilabelResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b",job="test"} 3
`
	labelFields1 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/org/:orgId"} 1
ui_page_render_errors{path="/prom/:orgId"} 1
`
	labelFields2 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/prom/:orgId"} 1
`
	labelFieldResult = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{job="test",path="/org/:orgId"} 1
ui_page_render_errors{job="test",path="/prom/:orgId"} 2
`
	gaugeInput = `
# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="ga",loaded="true"} 1
ui_external_lib_loaded{name="Intercom",loaded="true"} 1
ui_external_lib_loaded{name="mixpanel",loaded="true"} 1
`
	gaugeOutput = `# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{job="test",loaded="true",name="Intercom"} 2
ui_external_lib_loaded{job="test",loaded="true",name="ga"} 2
ui_external_lib_loaded{job="test",loaded="true",name="mixpanel"} 2
`

	gaugeBatchedOutput = `# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{job="test",loaded="true",name="Intercom"} 1
ui_external_lib_loaded{job="test",loaded="true",name="ga"} 1
ui_external_lib_loaded{job="test",loaded="true",name="mixpanel"} 1
`
	duplicateLabels = `
# HELP ui_external_lib_loaded Test with duplicate values
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="Munchkin",loaded="true"} 15171
ui_external_lib_loaded{name="Munchkin",loaded="true"} 1
`
	duplicateError = `duplicate labels: {__name__="ui_external_lib_loaded", job="test", loaded="true", name="Munchkin"}`

	reorderedLabels1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 1
`
	reorderedLabels2 = `# HELP counter A counter
# TYPE counter counter
counter{b="b",a="a"} 2
`
	reorderedLabelsResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b",job="test"} 3
`

	ignoredLabels1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b",ignore_me="ignored"} 1
`
	ignoredLabels2 = `# HELP counter A counter
# TYPE counter counter
counter{b="b",a="a",ignore_me="ignored"} 2
`
	ignoredLabelsResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b",job="test"} 3
`
)

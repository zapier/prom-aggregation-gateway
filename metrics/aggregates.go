package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/ulule/deepcopier"
)

const (
	labelParam = "labels"
)

// write to one read from the other than swap on a interval
type Aggregates struct {
	writeAggregate *Aggregate
	readAggregate  *Aggregate
	stop           chan bool
	batchTime      time.Duration
	ticker         *time.Ticker
}

func NewAggregates(metricBatchTime time.Duration, opts ...aggregateOptionsFunc) *Aggregates {
	a := &Aggregates{
		writeAggregate: newAggregate(opts...),
		readAggregate:  newAggregate(opts...),
		stop:           make(chan bool),
		batchTime:      metricBatchTime,
	}
	a.startAggregatesCopyBatch()

	return a
}

func (as *Aggregates) HandleInsert(c *gin.Context) {
	as.writeAggregate.handleInsert(c)
}

func (as *Aggregates) HandleRender(c *gin.Context) {
	as.readAggregate.handleRender(c)
}

func (as *Aggregates) Stop() {
	as.ticker.Stop()
	close(as.stop)
}

func (as *Aggregates) copyAggregateData() {
	as.writeAggregate.familiesLock.Lock()
	defer as.writeAggregate.familiesLock.Unlock()

	as.readAggregate.familiesLock.Lock()
	defer as.readAggregate.familiesLock.Unlock()

	for k, v := range as.writeAggregate.families {

		listOfMetric := []*dto.Metric{}
		for _, m := range v.Metric {
			// TODO: add logic to filter out old metrics here
			// if the time isn't pass the TTL time
			// if time.Since(time.UnixMilli(*m.TimestampMs)) < *as.writeAggregate.options.metricTTLDuration || as.writeAggregate.options.metricTTLDuration != nil {
			nMetric := &dto.Metric{}
			_ = deepcopier.Copy(m).To(nMetric)
			listOfMetric = append(listOfMetric, nMetric)
			// }
			// TODO: remove metric from writeAggregate as this could be a memory leak
		}

		as.readAggregate.families[k] = &metricFamily{

			MetricFamily: &dto.MetricFamily{
				Name:   v.Name,
				Help:   v.Help,
				Type:   v.Type,
				Metric: listOfMetric,
			},
		}
		if *v.Type == dto.MetricType_GAUGE {
			delete(as.writeAggregate.families, k)
		}
	}
}

func (as *Aggregates) startAggregatesCopyBatch() {
	as.ticker = time.NewTicker(as.batchTime)
	go func() {
		for {
			select {
			case <-as.stop:
				return
			case <-as.ticker.C:
				as.copyAggregateData()
			}
		}
	}()
}

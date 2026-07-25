package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"enterprise-search/internal/analytics"
	"google.golang.org/api/iterator"
)

// BigQueryClient defines the interface for the BigQuery client to allow mocking.
type BigQueryClient interface {
	Dataset(id string) DatasetHandle
	Query(q string) QueryHandle
	Close() error
}

type DatasetHandle interface {
	Table(id string) TableHandle
}

type TableHandle interface {
	Inserter() InserterHandle
}

type InserterHandle interface {
	Put(ctx context.Context, src any) error
}

type QueryHandle interface {
	Read(ctx context.Context) (RowIterator, error)
	SetParameters(params []bigquery.QueryParameter)
}

type RowIterator interface {
	Next(dst any) error
}

// Options controls construction of the BigQueryAdapter.
type Options struct {
	ProjectID     string
	DatasetID     string
	TableID       string
	ClientFactory func(ctx context.Context, projectID string) (BigQueryClient, error)
}

// BigQueryAdapter implements analytics.EventStore using Google Cloud BigQuery.
type BigQueryAdapter struct {
	client    BigQueryClient
	datasetID string
	tableID   string
}

var _ analytics.EventStore = (*BigQueryAdapter)(nil)

// NewBigQueryAdapter constructs a new BigQueryAdapter with default options.
func NewBigQueryAdapter(ctx context.Context, projectID, datasetID, tableID string) (*BigQueryAdapter, error) {
	return NewBigQueryAdapterWithOptions(ctx, Options{
		ProjectID: projectID,
		DatasetID: datasetID,
		TableID:   tableID,
	})
}

// NewBigQueryAdapterWithOptions constructs a new BigQueryAdapter with custom options.
func NewBigQueryAdapterWithOptions(ctx context.Context, opts Options) (*BigQueryAdapter, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if opts.DatasetID == "" {
		return nil, fmt.Errorf("dataset ID is required")
	}
	if opts.TableID == "" {
		return nil, fmt.Errorf("table ID is required")
	}

	if opts.ClientFactory == nil {
		opts.ClientFactory = func(ctx context.Context, projectID string) (BigQueryClient, error) {
			client, err := bigquery.NewClient(ctx, projectID)
			if err != nil {
				return nil, err
			}
			return &sdkClient{client: client}, nil
		}
	}

	client, err := opts.ClientFactory(ctx, opts.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("initialize BigQuery client: %w", err)
	}

	return &BigQueryAdapter{
		client:    client,
		datasetID: opts.DatasetID,
		tableID:   opts.TableID,
	}, nil
}

// Close closes the underlying BigQuery client.
func (a *BigQueryAdapter) Close() error {
	return a.client.Close()
}

type eventRow struct {
	ID        string    `bigquery:"id"`
	Name      string    `bigquery:"name"`
	Timestamp time.Time `bigquery:"timestamp"`
	Payload   string    `bigquery:"payload"`
}

// Write inserts an event into the BigQuery table.
func (a *BigQueryAdapter) Write(ctx context.Context, event *analytics.Event) error {
	if event == nil {
		return analytics.ErrInvalidEvent
	}
	if event.ID == "" || event.Name == "" {
		return analytics.ErrInvalidEvent
	}

	row := &eventRow{
		ID:        event.ID,
		Name:      event.Name,
		Timestamp: event.Timestamp,
		Payload:   event.Payload,
	}

	inserter := a.client.Dataset(a.datasetID).Table(a.tableID).Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		return fmt.Errorf("insert event to BigQuery: %w", err)
	}
	return nil
}

// Query retrieves events from the BigQuery table matching the filter.
func (a *BigQueryAdapter) Query(ctx context.Context, filter analytics.QueryFilter) ([]*analytics.Event, error) {
	queryStr := fmt.Sprintf("SELECT id, name, timestamp, payload FROM `%s.%s` WHERE 1=1", a.datasetID, a.tableID)
	var params []bigquery.QueryParameter

	if filter.Name != "" {
		queryStr += " AND name = @name"
		params = append(params, bigquery.QueryParameter{
			Name:  "name",
			Value: filter.Name,
		})
	}
	if !filter.StartTime.IsZero() {
		queryStr += " AND timestamp >= @start_time"
		params = append(params, bigquery.QueryParameter{
			Name:  "start_time",
			Value: filter.StartTime,
		})
	}
	if !filter.EndTime.IsZero() {
		queryStr += " AND timestamp <= @end_time"
		params = append(params, bigquery.QueryParameter{
			Name:  "end_time",
			Value: filter.EndTime,
		})
	}

	queryStr += " ORDER BY timestamp ASC"

	q := a.client.Query(queryStr)
	if len(params) > 0 {
		q.SetParameters(params)
	}

	iter, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("execute BigQuery query: %w", err)
	}

	var events []*analytics.Event
	for {
		var row eventRow
		err := iter.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read BigQuery row: %w", err)
		}

		events = append(events, &analytics.Event{
			ID:        row.ID,
			Name:      row.Name,
			Timestamp: row.Timestamp,
			Payload:   row.Payload,
		})
	}

	return events, nil
}

// SDK wrappers implementing the BigQueryClient interfaces.

type sdkClient struct {
	client *bigquery.Client
}

func (c *sdkClient) Dataset(id string) DatasetHandle {
	return &sdkDataset{dataset: c.client.Dataset(id)}
}

func (c *sdkClient) Query(q string) QueryHandle {
	return &sdkQuery{query: c.client.Query(q)}
}

func (c *sdkClient) Close() error {
	return c.client.Close()
}

type sdkDataset struct {
	dataset *bigquery.Dataset
}

func (d *sdkDataset) Table(id string) TableHandle {
	return &sdkTable{table: d.dataset.Table(id)}
}

type sdkTable struct {
	table *bigquery.Table
}

func (t *sdkTable) Inserter() InserterHandle {
	return &sdkInserter{inserter: t.table.Inserter()}
}

type sdkInserter struct {
	inserter *bigquery.Inserter
}

func (i *sdkInserter) Put(ctx context.Context, src any) error {
	return i.inserter.Put(ctx, src)
}

type sdkQuery struct {
	query *bigquery.Query
}

func (q *sdkQuery) Read(ctx context.Context) (RowIterator, error) {
	iter, err := q.query.Read(ctx)
	if err != nil {
		return nil, err
	}
	return &sdkRowIterator{iter: iter}, nil
}

func (q *sdkQuery) SetParameters(params []bigquery.QueryParameter) {
	q.query.Parameters = params
}

type sdkRowIterator struct {
	iter *bigquery.RowIterator
}

func (r *sdkRowIterator) Next(dst any) error {
	return r.iter.Next(dst)
}

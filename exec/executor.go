package exec

import (
	"context"
	"fmt"

	// "reflect"

	// "os"
	// "runtime/pprof"
	"cloud.google.com/go/bigquery"
	// "github.com/99designs/gqlgen/graphql"
	"github.com/HC-IPPM/cloud-cost-api/graph/model"
	"google.golang.org/api/iterator"
)

func parseNullFloat64(nullFloat bigquery.NullFloat64) *float64 {
	var value float64
	if nullFloat.Valid {
		value = nullFloat.Float64
	}
	return &value
}

func parseNullString(nullString bigquery.NullString) *string {
	var value string
	if nullString.Valid {
		value = nullString.StringVal
	}
	return &value
}

func query(ctx context.Context, client *bigquery.Client, project_ids_ptr []*string, needAllProjects bool) ([]*model.Project, error) {

	queryString := CostQueryStr(needAllProjects)

	project_ids := []string{}
	for _, el := range project_ids_ptr {
		project_ids = append(project_ids, *el)
	}
	q := client.Query(queryString)
	if !needAllProjects {
		q.Parameters = []bigquery.QueryParameter{
			{Name: "project_ids", Value: project_ids},
		}
	}

	q.Location = "northamerica-northeast1"

	job, err := q.Run(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	status, err := job.Wait(ctx)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if err := status.Err(); err != nil {
		return nil, err
	}

	it, err := job.Read(ctx)
	var projects []*model.Project
	for {
		var tempStruct struct {
			ProjectID                        bigquery.NullString    `bigquery:"project_id"`
			CurrentMonthToDate               bigquery.NullFloat64   `bigquery:"currentMonthToDate"`
			PreviousMonth                    bigquery.NullFloat64   `bigquery:"previousMonth"`
			CurrentMonthDeltaPercentage      bigquery.NullFloat64   `bigquery:"currentMonthDeltaPercentage"`
			CurrentCalendarYearToDate        bigquery.NullFloat64   `bigquery:"currentCalendarYearToDate"`
			PreviousCalendarYear             bigquery.NullFloat64   `bigquery:"previousCalendarYear"`
			CurrentFiscalToDate              bigquery.NullFloat64   `bigquery:"currentFiscalToDate"`
			PreviousFiscalYear               bigquery.NullFloat64   `bigquery:"previousFiscalYear"`
			CurrentFiscalYearDeltaPercentage bigquery.NullFloat64   `bigquery:"currentFiscalYearDeltaPercentage"`
			LastSixMonths                    []bigquery.NullFloat64 `bigquery:"lastSixMonths"`
		}

		var project model.Project
		var costs model.Cost

		err := it.Next(&tempStruct)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		project.Costs = &costs
		project.ID = *parseNullString(tempStruct.ProjectID)
		costs.CurrentMonthToDate = parseNullFloat64(tempStruct.CurrentMonthToDate)
		costs.PreviousMonth = parseNullFloat64(tempStruct.PreviousMonth)
		costs.CurrentMonthDeltaPercentage = parseNullFloat64(tempStruct.CurrentMonthDeltaPercentage)
		costs.CurrentCalendarYearToDate = parseNullFloat64(tempStruct.CurrentCalendarYearToDate)
		costs.PreviousCalendarYear = parseNullFloat64(tempStruct.PreviousCalendarYear)
		costs.CurrentFiscalToDate = parseNullFloat64(tempStruct.CurrentFiscalToDate)
		costs.PreviousFiscalYear = parseNullFloat64(tempStruct.PreviousFiscalYear)
		costs.CurrentFiscalYearDeltaPercentage = parseNullFloat64(tempStruct.CurrentFiscalYearDeltaPercentage)
		for _, cost := range tempStruct.LastSixMonths {
			project.Costs.LastSixMonths = append(project.Costs.LastSixMonths, parseNullFloat64(cost))
		}
		projects = append(projects, &project)
	}
	return projects, nil
}

func ExecuteQuery(ctx context.Context, client *bigquery.Client, project_ids []*string, needAllProjects bool) ([]*model.Project, error) {
	return query(ctx, client, project_ids, needAllProjects)
}

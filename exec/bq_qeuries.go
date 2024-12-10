package exec

func CostQueryStr(needAllProjects bool) string {
	var project_id_string string
	if !needAllProjects {
		project_id_string = `WHERE project.id IN UNNEST(@project_ids)`
	}

	return `
    WITH 
      today_date AS (SELECT CURRENT_DATE()),
      current_year AS (SELECT EXTRACT (year from (SELECT * FROM today_date)) as year),
      fiscal_year AS (
        SELECT 
        CASE 
          WHEN EXTRACT (month FROM (SELECT * FROM today_date)) > 3 THEN STRUCT(DATE(CONCAT(CAST(current_year.year -1 as string),'-04-01')) AS prev_fy_start,
                                                                              DATE(CONCAT(CAST(current_year.year    as string),'-03-31')) AS prev_fy_end,
                                                                              DATE(CONCAT(CAST(current_year.year    as string),'-04-01')) AS current_fy_start, 
                                                                              DATE(CONCAT(CAST(current_year.year +1 as string),'-03-31')) AS current_fy_end
                                                                  )
          WHEN EXTRACT (month FROM (SELECT * FROM today_date)) <= 3 THEN STRUCT(DATE(CONCAT(CAST(current_year.year - 2 as string),'-04-01')) AS prev_fy_start,
                                                                                DATE(CONCAT(CAST(current_year.year - 1 as string),'-03-31')) AS prev_fy_end,
                                                                                DATE(CONCAT(CAST(current_year.year - 1 as string),'-04-01')) AS current_fy_start, 
                                                                                DATE(CONCAT(CAST(current_year.year     as string),'-03-31')) AS current_fy_end
                                                                    )
        END as fy
      FROM current_year
      ),
      billing_data AS (
        SELECT
        TIMESTAMP_TRUNC(usage_start_time, MONTH) AS month,
        EXTRACT(YEAR FROM usage_start_time) AS year,
        project.id AS project_id,
    
        -- Cost with no credits applied
        SUM(cost) AS final_cost_no_credits,
    
        SUM(CASE 
            WHEN ARRAY_LENGTH(credits) = 0 THEN 0 
            ELSE credits[0].amount END) 
        as total_credits_computed1,
    
        SUM(COALESCE((SELECT SUM(x.amount) FROM UNNEST(credits) x),0))
        as total_credits_computed2,
        -- SUM(credits_sum_amount) AS total_credits_computed2,
    
        -- MY LOGIC TO COMPUTE FINAL_COST
        SUM(cost) + SUM(CASE 
          WHEN ARRAY_LENGTH(credits) = 0 THEN 0 
          ELSE credits[0].amount END) 
        as final_cost_0,
    
        -- billing_export_view LOGIC TO COMPUTE FINAL_COST, the credits_sum_amount is computed with COALESCE((SELECT SUM(x.amount) FROM UNNEST(credits) x),0)
        SUM(cost) + SUM(credits_sum_amount) 
        as final_cost
    
      FROM` + "`pdcp-serv-001-budgets.billing_daily_costs.billing-export-view` " + project_id_string +
		`GROUP BY 1,2,3
      ORDER BY project_id, month DESC
    ),
    project_ids AS (
      SELECT project_id
      FROM billing_data
      GROUP BY project_id
    ),
    current_month AS (
      SELECT
        final_cost as cost, project_id
      FROM billing_data
      WHERE DATE(month) = TIMESTAMP_TRUNC(CURRENT_DATE(), MONTH)
    ),
    previous_month AS (
      SELECT
        final_cost as cost, project_id
      FROM billing_data
      WHERE DATE(month) = TIMESTAMP_TRUNC(DATE_SUB(CURRENT_DATE(), INTERVAL 1 MONTH), MONTH)
    ),
    last_six_months AS (
      SELECT
        ARRAY_AGG(final_cost) as lsm_cost, project_id
      FROM billing_data
      WHERE DATE(month) >= TIMESTAMP_TRUNC(DATE_SUB(CURRENT_DATE(), INTERVAL 6 MONTH), MONTH) AND DATE(month) < TIMESTAMP_TRUNC(CURRENT_DATE(), MONTH)
      GROUP BY project_id
    ),
    current_calendar_year AS (
      SELECT
        SUM(final_cost) AS cost, project_id
      FROM billing_data
      WHERE year = EXTRACT(YEAR FROM CURRENT_DATE())
      GROUP BY project_id
    ),
    previous_calendar_year AS (
      SELECT
        sum(final_cost) as cost, project_id
      FROM billing_data
      WHERE year = EXTRACT(YEAR FROM DATE_SUB(CURRENT_DATE(), INTERVAL 1 YEAR))
      GROUP BY project_id
    ),
    current_fiscal_year AS (
      SELECT
        SUM(final_cost) AS cost, project_id
      FROM billing_data
      WHERE DATE(month) >= (SELECT fy.current_fy_start FROM fiscal_year) AND DATE(month) <= (SELECT fy.current_fy_end FROM fiscal_year)
      GROUP BY project_id
    ),
    previous_fiscal_year AS (
      SELECT
        SUM(final_cost) AS cost, project_id
      FROM billing_data
      WHERE DATE(month) >= (SELECT fy.prev_fy_start FROM fiscal_year) AND DATE(month) <= (SELECT fy.prev_fy_end FROM fiscal_year)
      GROUP BY project_id
    )
    SELECT
    pid.project_id,
    cm.cost as currentMonthToDate,
    pm.cost as previousMonth,
    (SAFE_DIVIDE(cm.cost - pm.cost, pm.cost) * 100) as currentMonthDeltaPercentage,
    ccy.cost as currentCalendarYearToDate,
    pcy.cost as previousCalendarYear,
    cfy.cost as currentFiscalToDate,
    pfy.cost as previousFiscalYear,
    (SAFE_DIVIDE(cfy.cost - pfy.cost , pfy.cost) * 100) as currentFiscalYearDeltaPercentage,
    lsm.lsm_cost as lastSixMonths
    FROM
    project_ids pid
    LEFT JOIN current_month cm ON pid.project_id = cm.project_id
    LEFT JOIN previous_month pm ON pid.project_id = pm.project_id
    LEFT JOIN previous_calendar_year pcy  ON pid.project_id = pcy.project_id
    LEFT JOIN current_calendar_year ccy  ON pid.project_id = ccy.project_id
    LEFT JOIN current_fiscal_year cfy  ON pid.project_id = cfy.project_id
    LEFT JOIN previous_fiscal_year pfy  ON pid.project_id = pfy.project_id
    LEFT JOIN last_six_months lsm  ON pid.project_id = lsm.project_id
    `
}

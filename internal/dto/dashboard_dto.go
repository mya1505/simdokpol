package dto

type DashboardStatsDTO struct {
	DocsMonthly int64 `json:"docs_monthly"`
	DocsYearly  int64 `json:"docs_yearly"`
	DocsToday   int64 `json:"docs_today"`
	ActiveUsers int64 `json:"active_users"`
}

type ChartDataDTO struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}

type PieChartDataDTO struct {
	Labels           []string `json:"labels"`
	Data             []int    `json:"data"`
	BackgroundColors []string `json:"background_colors"`
}
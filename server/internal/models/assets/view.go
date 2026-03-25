package assets

// AssetOSTypeStats 系统类型统计视图
type AssetOSTypeStats struct {
	OSFamily string `json:"os_family" gorm:"column:os_family"`
	Count    int64  `json:"count" gorm:"column:count"`
}

// TableName 指定表名
func (AssetOSTypeStats) TableName() string {
	return "v_asset_ostype_stats"
}

// AssetHostStats 主机统计视图
type AssetHostStats struct {
	TodayTotal       int64   `json:"today_total" gorm:"column:today_total"`
	YesterdayTotal   int64   `json:"yesterday_total" gorm:"column:yesterday_total"`
	NetIncrease      int64   `json:"net_increase" gorm:"column:net_increase"`
	GrowthPercentage float64 `json:"growth_percentage" gorm:"column:growth_percentage"`
}

// TableName 指定表名
func (AssetHostStats) TableName() string {
	return "v_asset_host_stats"
}

// AssetDatabaseTypeStats 数据库类型统计视图
type AssetDatabaseTypeStats struct {
	DBTypeName string `json:"db_type_name" gorm:"column:db_type_name"`
	Count      int64  `json:"count" gorm:"column:count"`
}

// TableName 指定表名
func (AssetDatabaseTypeStats) TableName() string {
	return "v_asset_databasetype_stats"
}

// AssetDatabaseStats 数据库统计视图
type AssetDatabaseStats struct {
	TodayTotal       int64   `json:"today_total" gorm:"column:today_total"`
	YesterdayTotal   int64   `json:"yesterday_total" gorm:"column:yesterday_total"`
	NetIncrease      int64   `json:"net_increase" gorm:"column:net_increase"`
	GrowthPercentage float64 `json:"growth_percentage" gorm:"column:growth_percentage"`
}

// TableName 指定表名
func (AssetDatabaseStats) TableName() string {
	return "v_asset_database_stats"
}

// AssetContainerStats 容器统计视图
type AssetContainerStats struct {
	TodayTotal       int64   `json:"today_total" gorm:"column:today_total"`
	YesterdayTotal   int64   `json:"yesterday_total" gorm:"column:yesterday_total"`
	NetIncrease      int64   `json:"net_increase" gorm:"column:net_increase"`
	GrowthPercentage float64 `json:"growth_percentage" gorm:"column:growth_percentage"`
}

// TableName 指定表名
func (AssetContainerStats) TableName() string {
	return "v_asset_container_stats"
}

// AssetAccountStats 账号统计视图
type AssetAccountStats struct {
	TodayTotal       int64   `json:"today_total" gorm:"column:today_total"`
	YesterdayTotal   int64   `json:"yesterday_total" gorm:"column:yesterday_total"`
	NetIncrease      int64   `json:"net_increase" gorm:"column:net_increase"`
	GrowthPercentage float64 `json:"growth_percentage" gorm:"column:growth_percentage"`
}

// TableName 指定表名
func (AssetAccountStats) TableName() string {
	return "v_asset_account_stats"
}

// AssetLatestAssetsTop5 近期更新资产视图
type AssetLatestAssetsTop5 struct {
	ID          int64  `json:"id" gorm:"column:id"`
	AgentID     string `json:"agent_id" gorm:"column:agent_id"`
	SourceTable string `json:"source_table" gorm:"column:source_table"`
	TableNameCN string `json:"table_name_cn" gorm:"column:table_name_cn"`
	CreatedAt   string `json:"created_at" gorm:"column:created_at"`
}

// TableName 指定表名
func (AssetLatestAssetsTop5) TableName() string {
	return "v_asset_latest_assets_top5"
}

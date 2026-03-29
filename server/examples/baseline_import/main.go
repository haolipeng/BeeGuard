package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ===================== 配置结构 =====================

// DBConfig 数据库配置
type DBConfig struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"database"`
}

// ===================== YAML 数据结构 =====================

// BaselineYAML 1400.yaml 顶层结构
type BaselineYAML struct {
	BaselineID      int         `yaml:"baseline_id"`
	BaselineVersion string      `yaml:"baseline_version"`
	BaselineName    string      `yaml:"baseline_name"`
	TemplateID      int         `yaml:"template_id"`
	System          []string    `yaml:"system"`
	CheckList       []CheckItem `yaml:"check_list"`
}

// CheckItem check_list 中每条规则
type CheckItem struct {
	CheckID       int        `yaml:"check_id"`
	Type          string     `yaml:"type"`
	Title         string     `yaml:"title"`
	Description   string     `yaml:"description"`
	Solution      string     `yaml:"solution"`
	Security      string     `yaml:"security"`
	TypeCN        string     `yaml:"type_cn"`
	TitleCN       string     `yaml:"title_cn"`
	DescriptionCN string     `yaml:"description_cn"`
	SolutionCN    string     `yaml:"solution_cn"`
	Check         CheckRules `yaml:"check"`
}

// CheckRules check 字段结构
type CheckRules struct {
	Condition string     `yaml:"condition,omitempty" json:"condition,omitempty"`
	Rules     []RuleItem `yaml:"rules" json:"rules"`
}

// RuleItem 单条检查规则
type RuleItem struct {
	Type   string   `yaml:"type" json:"type"`
	Param  []string `yaml:"param" json:"param"`
	Filter string   `yaml:"filter,omitempty" json:"filter,omitempty"`
	Result string   `yaml:"result" json:"result"`
}

// ===================== 数据库模型 =====================

// BaselineTemplate 基线模板
type BaselineTemplate struct {
	ID           int64     `gorm:"primaryKey;not null"`
	TemplateName string    `gorm:"not null"`
	TemplateType string    `gorm:"not null"`
	OSType       *string   `gorm:"column:os_type"`
	Version      *string   `gorm:"column:version"`
	ItemCount    *int32    `gorm:"column:item_count"`
	Description  *string   `gorm:"column:description"`
	IsEnabled    int16     `gorm:"not null;default:1"`
	BaselineIDs  *string   `gorm:"type:text;column:baseline_ids"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (BaselineTemplate) TableName() string { return "baseline_template" }

// BaselineCheckItem 基线检查项
type BaselineCheckItem struct {
	ID            int64     `gorm:"primaryKey;not null"`
	TemplateID    int64     `gorm:"not null;index"`
	ItemName      string    `gorm:"not null"`
	Category      string    `gorm:"not null"`
	RiskLevel     string    `gorm:"not null"`
	CheckRules    string    `gorm:"not null"`
	FixSuggestion string    `gorm:"not null"`
	FixScript     string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (BaselineCheckItem) TableName() string { return "baseline_check_item" }

// BaselineTemplateHostLink 基线模板与主机关联
type BaselineTemplateHostLink struct {
	ID            int64     `gorm:"primaryKey;not null"`
	TemplateID    int64     `gorm:"column:baseline_template_id;not null;index"`
	TemplateName  string    `gorm:"not null"`
	TargetRange   *string   `gorm:"not null;type:text"`
	ScanFrequency string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (BaselineTemplateHostLink) TableName() string { return "baseline_template_host_link" }

// ===================== 主程序 =====================

func main() {
	configPath := flag.String("config", "config.yaml", "数据库配置文件路径")
	yamlPath := flag.String("yaml", "/home/work/goProject/src/BeeGuard/agent/business_plugins/baseline/config/linux/1400.yaml", "基线YAML文件路径")
	flag.Parse()

	// 1. 加载数据库配置
	dbCfg, err := loadDBConfig(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 连接数据库
	db, err := connectDB(dbCfg)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	fmt.Println("数据库连接成功")

	// 3. 解析 YAML
	baseline, err := loadBaselineYAML(*yamlPath)
	if err != nil {
		fmt.Printf("解析YAML失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已加载基线: %s (ID=%d, 版本=%s, 检查项=%d条)\n",
		baseline.BaselineName, baseline.BaselineID, baseline.BaselineVersion, len(baseline.CheckList))

	// 4. 交互式选择并写入
	insertedIDs, stats := interactiveImport(db, baseline)

	// 5. 更新 baseline_template
	updateTemplate(db, baseline, insertedIDs)

	// 6. 更新 baseline_template_host_link
	updateTemplateHostLink(db, baseline)

	// 7. 打印统计
	fmt.Println("\n========== 导入统计 ==========")
	fmt.Printf("总规则数:   %d\n", stats.total)
	fmt.Printf("新增写入:   %d\n", stats.inserted)
	fmt.Printf("已存在跳过: %d\n", stats.existed)
	fmt.Printf("用户跳过:   %d\n", stats.skipped)
	fmt.Println("==============================")
}

// ===================== 辅助函数 =====================

func loadDBConfig(path string) (*DBConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg DBConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

func connectDB(cfg *DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Database, cfg.Database.Port)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func loadBaselineYAML(path string) (*BaselineYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取YAML文件失败: %w", err)
	}
	var baseline BaselineYAML
	if err := yaml.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("解析YAML文件失败: %w", err)
	}
	return &baseline, nil
}

type importStats struct {
	total    int
	inserted int
	existed  int
	skipped  int
}

func interactiveImport(db *gorm.DB, baseline *BaselineYAML) ([]int64, importStats) {
	scanner := bufio.NewScanner(os.Stdin)
	var insertedIDs []int64
	var stats importStats
	allMode := false
	baselineID := int64(baseline.BaselineID)

	stats.total = len(baseline.CheckList)

	for i, item := range baseline.CheckList {
		fmt.Printf("\n[%d/%d] check_id=%d | %s | %s | %s\n",
			i+1, len(baseline.CheckList), item.CheckID, item.TitleCN, item.TypeCN, item.Security)

		if !allMode {
			fmt.Print("写入此规则? (y=写入 / n=跳过 / a=全部写入 / q=退出): ")
			scanner.Scan()
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			switch input {
			case "n":
				stats.skipped++
				continue
			case "q":
				stats.skipped += len(baseline.CheckList) - i
				return insertedIDs, stats
			case "a":
				allMode = true
			case "y":
				// 继续写入
			default:
				fmt.Println("无效输入，跳过此项")
				stats.skipped++
				continue
			}
		}

		// 检查是否已存在 (按 template_id + item_name 去重)
		var existing BaselineCheckItem
		result := db.Where("template_id = ? AND item_name = ?", baselineID, item.TitleCN).First(&existing)
		if result.Error == nil {
			fmt.Printf("  -> 已存在 (id=%d)，跳过\n", existing.ID)
			insertedIDs = append(insertedIDs, existing.ID)
			stats.existed++
			continue
		}

		// 序列化 check 字段为 JSON
		checkJSON, err := json.Marshal(item.Check)
		if err != nil {
			fmt.Printf("  -> JSON序列化失败: %v，跳过\n", err)
			stats.skipped++
			continue
		}

		record := BaselineCheckItem{
			TemplateID:    baselineID,
			ItemName:      item.TitleCN,
			Category:      item.TypeCN,
			RiskLevel:     item.Security,
			CheckRules:    string(checkJSON),
			FixSuggestion: item.SolutionCN,
			FixScript:     "",
		}

		if err := db.Create(&record).Error; err != nil {
			fmt.Printf("  -> 写入失败: %v\n", err)
			stats.skipped++
			continue
		}

		fmt.Printf("  -> 写入成功 (id=%d)\n", record.ID)
		insertedIDs = append(insertedIDs, record.ID)
		stats.inserted++
	}

	return insertedIDs, stats
}

func updateTemplate(db *gorm.DB, baseline *BaselineYAML, insertedIDs []int64) {
	baselineID := int64(baseline.BaselineID)
	osType := "linux"
	version := baseline.BaselineVersion
	itemCount := int32(len(insertedIDs))
	description := baseline.BaselineName

	// 构建 baseline_ids 逗号分隔字符串
	var idStrs []string
	for _, id := range insertedIDs {
		idStrs = append(idStrs, strconv.FormatInt(id, 10))
	}
	baselineIDsStr := strings.Join(idStrs, ",")

	var existing BaselineTemplate
	result := db.Where("id = ?", baselineID).First(&existing)
	if result.Error != nil {
		// 不存在，插入
		tpl := BaselineTemplate{
			ID:           baselineID,
			TemplateName: baseline.BaselineName,
			TemplateType: "os_security",
			OSType:       &osType,
			Version:      &version,
			ItemCount:    &itemCount,
			Description:  &description,
			IsEnabled:    1,
			BaselineIDs:  &baselineIDsStr,
		}
		if err := db.Create(&tpl).Error; err != nil {
			fmt.Printf("创建 baseline_template 失败: %v\n", err)
			return
		}
		fmt.Printf("\n已创建 baseline_template (id=%d)\n", baselineID)
	} else {
		// 已存在，更新
		updates := map[string]interface{}{
			"template_name": baseline.BaselineName,
			"template_type": "os_security",
			"os_type":       osType,
			"version":       version,
			"item_count":    itemCount,
			"description":   description,
			"is_enabled":    1,
			"baseline_ids":  baselineIDsStr,
		}
		if err := db.Model(&BaselineTemplate{}).Where("id = ?", baselineID).Updates(updates).Error; err != nil {
			fmt.Printf("更新 baseline_template 失败: %v\n", err)
			return
		}
		fmt.Printf("\n已更新 baseline_template (id=%d)\n", baselineID)
	}
}

func updateTemplateHostLink(db *gorm.DB, baseline *BaselineYAML) {
	templateID := int64(baseline.TemplateID)
	if templateID == 0 {
		templateID = int64(baseline.BaselineID)
	}

	targetRange := "[]"

	// 检查是否已存在该模板的 host_link 记录
	var existing BaselineTemplateHostLink
	result := db.Where("baseline_template_id = ?", templateID).First(&existing)
	if result.Error == nil {
		// 已存在，更新
		updates := map[string]interface{}{
			"baseline_template_name": baseline.BaselineName,
			"target_range":           targetRange,
			"scan_frequency":         "daily",
		}
		if err := db.Model(&BaselineTemplateHostLink{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			fmt.Printf("更新 baseline_template_host_link 失败: %v\n", err)
			return
		}
		fmt.Printf("已更新 baseline_template_host_link (id=%d)\n", existing.ID)
	} else {
		// 不存在，创建
		link := BaselineTemplateHostLink{
			TemplateID:    templateID,
			TemplateName:  baseline.BaselineName,
			TargetRange:   &targetRange,
			ScanFrequency: "daily",
		}
		if err := db.Create(&link).Error; err != nil {
			fmt.Printf("创建 baseline_template_host_link 失败: %v\n", err)
			return
		}
		fmt.Printf("已创建 baseline_template_host_link (id=%d)\n", link.ID)
	}
}

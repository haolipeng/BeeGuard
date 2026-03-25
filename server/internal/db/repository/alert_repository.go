package repository

import (
	"context"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/alert"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
	"gorm.io/gorm"
)

// AlertRepository 告警数据仓库
type AlertRepository struct{}

// NewAlertRepository 创建告警仓库实例
func NewAlertRepository() *AlertRepository {
	return &AlertRepository{}
}

// CreateBruteForceAlert 创建暴力破解告警记录（每次告警都是新记录）
func (r *AlertRepository) CreateBruteForceAlert(ctx context.Context, a *alert.BruteForce) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 暴力破解告警写入失败: %v", err)
	}
	return err
}

// CreateDangerousCommandAlert 创建高危命令告警记录
func (r *AlertRepository) CreateDangerousCommandAlert(ctx context.Context, a *alert.DangerousCommand) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 高危命令告警写入失败: %v", err)
	}
	return err
}

// CreateReverseShellAlert 创建反弹Shell告警记录
func (r *AlertRepository) CreateReverseShellAlert(ctx context.Context, a *alert.ReverseShell) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 反弹Shell告警写入失败: %v", err)
	}
	return err
}

// CreateAbnormalLoginAlert 创建异常登录告警记录
func (r *AlertRepository) CreateAbnormalLoginAlert(ctx context.Context, a *alert.AbnormalLogin) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 异常登录告警写入失败: %v", err)
	}
	return err
}

// CreatePrivilegeEscalationAlert 创建本地提权告警记录
func (r *AlertRepository) CreatePrivilegeEscalationAlert(ctx context.Context, a *alert.PrivilegeEscalation) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 本地提权告警写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateMaliciousRequestAlert 创建或更新恶意请求告警（支持聚合）
// 聚合维度：agent_id + (malicious_domain OR malicious_ip)
func (r *AlertRepository) CreateOrUpdateMaliciousRequestAlert(ctx context.Context, a *alert.MaliciousRequest) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	// 查询是否存在相同聚合键的记录
	var existingAlert alert.MaliciousRequest

	// 构建查询条件：agent_id + (malicious_domain OR malicious_ip)
	query := database.WithContext(ctx).Where("agent_id = ?", a.AgentID)

	if a.MaliciousDomain != "" {
		query = query.Where("malicious_domain = ?", a.MaliciousDomain)
	} else if a.MaliciousIP != nil {
		query = query.Where("malicious_ip = ?", *a.MaliciousIP)
	} else {
		log.Warnf("[AlertRepository] 恶意请求告警缺少domain和IP，跳过写入")
		return nil
	}

	result := query.First(&existingAlert)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Errorf("[AlertRepository] 查询恶意请求告警失败: %v", result.Error)
		return result.Error
	}

	if result.Error == gorm.ErrRecordNotFound {
		// 记录不存在，创建新记录
		a.RequestCount = 1
		if a.FirstRequestTime == nil {
			now := common.DateTime{Time: time.Now()}
			a.FirstRequestTime = &now
		}
		if a.LastRequestTime == nil {
			now := common.DateTime{Time: time.Now()}
			a.LastRequestTime = &now
		}

		err := database.WithContext(ctx).Create(a).Error
		if err != nil {
			log.Errorf("[AlertRepository] 恶意请求告警创建失败: %v", err)
		}
		return err
	}

	// 记录已存在，更新计数和最后请求时间
	updates := map[string]interface{}{
		"request_count":     gorm.Expr("request_count + ?", 1),
		"last_request_time": a.LastRequestTime,
		"updated_at":        common.DateTime{Time: time.Now()},
	}

	// 如果新告警有risk_description且现有记录为空，则更新
	if a.RiskDescription != nil && *a.RiskDescription != "" &&
		(existingAlert.RiskDescription == nil || *existingAlert.RiskDescription == "") {
		updates["risk_description"] = a.RiskDescription
	}

	err := database.WithContext(ctx).Model(&existingAlert).Updates(updates).Error
	if err != nil {
		log.Errorf("[AlertRepository] 恶意请求告警更新失败: %v", err)
	}
	return err
}

// CreateMalwareScanAlert 创建恶意文件扫描告警记录（DataType 6061/6062）
func (r *AlertRepository) CreateMalwareScanAlert(ctx context.Context, a *alert.MalwareScan) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 恶意文件扫描告警写入失败: %v", err)
	}
	return err
}

// CreateFileIntegrityAlert 创建文件完整性告警记录
func (r *AlertRepository) CreateFileIntegrityAlert(ctx context.Context, a *alert.FileIntegrity) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 文件完整性告警写入失败: %v", err)
	}
	return err
}

// CreateNetworkAttackAlert 创建网络攻击告警记录
func (r *AlertRepository) CreateNetworkAttackAlert(ctx context.Context, a *alert.NetworkAttack) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 网络攻击告警写入失败: %v", err)
	}
	return err
}

// CreateContainerDangerousCommandAlert 创建容器高危命令告警记录
func (r *AlertRepository) CreateContainerDangerousCommandAlert(ctx context.Context, a *alert.ContainerDangerousCommand) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 容器高危命令告警写入失败: %v", err)
	}
	return err
}

// CreateContainerReverseShellAlert 创建容器反弹Shell告警记录
func (r *AlertRepository) CreateContainerReverseShellAlert(ctx context.Context, a *alert.ContainerReverseShell) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 容器反弹Shell告警写入失败: %v", err)
	}
	return err
}

// CreateContainerSensitiveFileAlert 创建容器核心文件监控告警记录
func (r *AlertRepository) CreateContainerSensitiveFileAlert(ctx context.Context, a *alert.ContainerSensitiveFile) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[AlertRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Errorf("[AlertRepository] 容器核心文件监控告警写入失败: %v", err)
	}
	return err
}

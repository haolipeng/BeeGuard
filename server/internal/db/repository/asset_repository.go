package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/container"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// AssetRepository 资产数据仓库
type AssetRepository struct{}

// NewAssetRepository 创建资产仓库实例
func NewAssetRepository() *AssetRepository {
	return &AssetRepository{}
}

// getDB 获取数据库连接
func (r *AssetRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// CreateOrUpdateHost 插入或更新主机记录
func (r *AssetRepository) CreateOrUpdateHost(ctx context.Context, hostObj *host.Host) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[Repository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "agent_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "mac_addr", "os_type", "os_version",
			"agent_status", "agent_version", "last_heartbeat", "updated_at",
		}),
	}).Create(hostObj).Error

	if err != nil {
		log.Errorf("[Repository] 主机写入失败: %v", err)
	}
	return err
}

// CreateOrUpdatePort 插入或更新端口记录
func (r *AssetRepository) CreateOrUpdatePort(ctx context.Context, port *host.Port) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "port"},
			{Name: "protocol"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "listen_ip", "listen_process",
			"run_user", "agent_status", "process_time", "updated_at",
		}),
	}).Create(port).Error

	if err != nil {
		log.Errorf("[Repository] 端口写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateAccount 插入或更新账号记录
func (r *AssetRepository) CreateOrUpdateAccount(ctx context.Context, account *host.Account) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "uid", "status", "permission",
			"login_type", "last_login_time", "updated_at",
		}),
	}).Create(account).Error

	if err != nil {
		log.Errorf("[Repository] 账号写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateProcess 插入或更新进程记录
func (r *AssetRepository) CreateOrUpdateProcess(ctx context.Context, process *host.Process) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "path"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "name", "status",
			"run_name", "start_time", "updated_at",
		}),
	}).Create(process).Error

	if err != nil {
		log.Errorf("[Repository] 进程写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateDatabase 插入或更新数据库服务记录
func (r *AssetRepository) CreateOrUpdateDatabase(ctx context.Context, database *host.Database) error {
	dbConn := r.getDB()
	if dbConn == nil {
		return nil
	}

	err := dbConn.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "db_type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "db_version", "port",
			"run_user", "updated_at",
		}),
	}).Create(database).Error

	if err != nil {
		log.Errorf("[Repository] 数据库服务写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateWebService 插入或更新Web服务记录
func (r *AssetRepository) CreateOrUpdateWebService(ctx context.Context, webService *host.Web) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "server_type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "name", "version",
			"site_domain", "path", "updated_at",
		}),
	}).Create(webService).Error

	if err != nil {
		log.Errorf("[Repository] Web服务写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateSystemService 插入或更新系统服务记录
func (r *AssetRepository) CreateOrUpdateSystemService(ctx context.Context, service *host.System) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "version", "status",
			"run_user", "path", "describe", "updated_at",
		}),
	}).Create(service).Error

	if err != nil {
		log.Errorf("[Repository] 系统服务写入失败: %v", err)
	}
	return err
}

// BatchCreateOrUpdatePorts 批量插入或更新端口记录
func (r *AssetRepository) BatchCreateOrUpdatePorts(ctx context.Context, ports []*host.Port) error {
	if len(ports) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "port"},
			{Name: "protocol"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "listen_ip", "listen_process",
			"run_user", "agent_status", "process_time", "updated_at",
		}),
	}).Create(&ports).Error

	if err != nil {
		log.Errorf("[Repository] 端口批量写入失败: %v", err)
	}
	return err
}

// BatchCreateOrUpdateAccounts 批量插入或更新账号记录
func (r *AssetRepository) BatchCreateOrUpdateAccounts(ctx context.Context, accounts []*host.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "uid", "status", "permission",
			"login_type", "last_login_time", "updated_at",
		}),
	}).Create(&accounts).Error

	if err != nil {
		log.Errorf("[Repository] 账号批量写入失败: %v", err)
	}
	return err
}

// BatchCreateOrUpdateProcesses 批量插入或更新进程记录
func (r *AssetRepository) BatchCreateOrUpdateProcesses(ctx context.Context, processes []*host.Process) error {
	if len(processes) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "path"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "name", "status",
			"run_name", "start_time", "updated_at",
		}),
	}).Create(&processes).Error

	if err != nil {
		log.Errorf("[Repository] 进程批量写入失败: %v", err)
	}
	return err
}

// UpdateHostOffline 更新主机离线状态
func (r *AssetRepository) UpdateHostOffline(ctx context.Context, agentID string) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).
		Model(&host.Host{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"agent_status": 0,
			"updated_at":   time.Now(),
		}).Error

	return err
}

// BatchCreateOrUpdateSoftware 批量插入或更新软件记录
func (r *AssetRepository) BatchCreateOrUpdateSoftware(ctx context.Context, softwares []*host.Software) error {
	if len(softwares) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "name"},
			{Name: "type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "version", "source",
			"status", "vendor", "path", "updated_at",
		}),
	}).Create(&softwares).Error

	if err != nil {
		log.Errorf("[Repository] 软件批量写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateContainer 插入或更新容器记录
func (r *AssetRepository) CreateOrUpdateContainer(ctx context.Context, container *container.Container) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "container_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "name", "state",
			"image_id", "image_name", "runtime", "pid",
			"create_time", "updated_at",
		}),
	}).Create(container).Error

	if err != nil {
		log.Errorf("[Repository] 容器写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateEnvSuspicious 插入或更新可疑环境变量记录
func (r *AssetRepository) CreateOrUpdateEnvSuspicious(ctx context.Context, env *host.EnvSuspicious) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "var_name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "var_value",
			"suspicious_reasons", "source", "updated_at",
		}),
	}).Create(env).Error

	if err != nil {
		log.Errorf("[Repository] 可疑环境变量写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateKmod 插入或更新内核模块记录
func (r *AssetRepository) CreateOrUpdateKmod(ctx context.Context, kmod *host.Kmod) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "size", "refcount",
			"used_by", "state", "addr", "updated_at",
		}),
	}).Create(kmod).Error

	if err != nil {
		log.Errorf("[Repository] 内核模块写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateImagePackage 插入或更新镜像软件包记录
func (r *AssetRepository) CreateOrUpdateImagePackage(ctx context.Context, pkg *container.ImagePackage) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "image_id"},
			{Name: "package_name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "image_name", "package_version",
			"package_type", "os_version", "updated_at",
		}),
	}).Create(pkg).Error

	if err != nil {
		log.Errorf("[Repository] 镜像软件包写入失败: %v", err)
	}
	return err
}

// CreateOrUpdateImage 插入或更新镜像记录
func (r *AssetRepository) CreateOrUpdateImage(ctx context.Context, image *container.Image) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
			{Name: "image_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"host_name", "host_ip", "image_name", "image_version",
			"image_size", "container_count", "build_time",
			"runtime", "updated_at",
		}),
	}).Create(image).Error

	if err != nil {
		log.Errorf("[Repository] 镜像写入失败: %v", err)
	}
	return err
}

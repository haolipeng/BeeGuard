package models

// InitMenus 初始化菜单结构
func InitMenus() error {
	// 由于用户不需要menus表，跳过初始化
	return nil
}

// AutoMigrate 迁移所有表结构
func AutoMigrate() error {
	// 注释掉自动迁移，因为我们已经有现成的数据库表结构
	/*
	db := mysql.DB

	err := db.AutoMigrate(
		&system.Menu{},
		&system.Client{},
		&system.Server{},
		&HostAsset{},
		&ContainerAsset{},
		//&code.CodeQlScanResult{}, // 添加CodeQL扫描结果模型
		&code.Repos{},             // 添加代码仓库模型
		&code.RepoScanList{},      // 添加仓库扫描列表模型
		&code.CodeqlScanResults{}, // 添加CodeQL扫描结果模型
		&code.Rules{},             // 添加规则集模型
		&back.CodeqlRule{},        // 添加规则详情模型
	)

	if err != nil {
		return err
	}
	*/

	// 初始化菜单
	err := InitMenus()
	if err != nil {
		return err
	}

	return nil
}
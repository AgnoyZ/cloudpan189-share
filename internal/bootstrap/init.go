package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	"go.uber.org/zap"

	"github.com/casbin/casbin/v2/model"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"github.com/xxcheng123/cloudpan189-share/internal/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func useSQLiteDB(c *configs.Config) (db *gorm.DB, err error) {
	dir := filepath.Dir(c.DBFile)
	if err = os.MkdirAll(dir, 0766); err != nil {
		return
	}

	db, err = gorm.Open(sqlite.Open(c.DBFile), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err = db.Use(new(TracePlugin)); err != nil {
		return nil, fmt.Errorf("failed to register trace plugin: %w", err)
	}

	return db, nil
}

func useMySqlDB(c *configs.Config) (db *gorm.DB, err error) {
	if c.MySQL == nil {
		return nil, fmt.Errorf("MySQL configuration is required when using MySQL database")
	}

	// 构建 MySQL DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.User,
		c.MySQL.Pass,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.DBName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
	}

	if err = db.Use(new(TracePlugin)); err != nil {
		return nil, fmt.Errorf("failed to register trace plugin: %w", err)
	}

	return db, nil
}

func connectDB(c *configs.Config) (db *gorm.DB, err error) {
	switch c.DBType {
	case "mysql":
		return useMySqlDB(c)
	case "sqlite":
		return useSQLiteDB(c)
	default:
		return useSQLiteDB(c) // 默认使用 SQLite
	}
}

func assignShared(db *gorm.DB) (err error) {
	var setting = new(models.Setting)
	if err = db.First(setting).Error; err != nil {
		return err
	}

	shared.SaltKey = setting.SaltKey
	shared.BaseURL = setting.BaseURL
	shared.SettingAddition = setting.Addition

	return nil
}

// initFileEnforcer 初始化Casbin文件权限执行器
func initFileEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, new(models.Group2File), new(models.UserGroup).TableName())
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	m, _ := model.NewModelFromString(fileModelTemplate)

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 加载策略
	if err = enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policy: %w", err)
	}

	return enforcer, nil
}

const fileModelTemplate = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || g(r.sub, "default_group")`

func initTaskEngine(logger *zap.Logger) taskengine.TaskEngine {
	return taskengine.NewTaskEngine(taskengine.WithLogger(logger.Named("task_engine")))
}

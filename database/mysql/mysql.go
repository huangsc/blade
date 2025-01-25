package mysql

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/huangsc/blade/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB MySQL数据库实现
type DB struct {
	db  *gorm.DB
	cfg *database.Config
}

// NewDB 创建MySQL数据库实例
func NewDB(opts ...database.Option) (database.DB, error) {
	// 创建默认配置
	cfg := &database.Config{
		DSN:             "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
		SlowThreshold:   time.Millisecond * 500,
		Logger:          os.Stdout,
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(cfg)
	}

	// 创建GORM配置
	gormCfg := &gorm.Config{
		Logger: logger.New(
			log.New(cfg.Logger, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             cfg.SlowThreshold,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 创建GORM实例
	gormDB, err := gorm.Open(mysql.Open(cfg.DSN), gormCfg)
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return &DB{
		db:  gormDB,
		cfg: cfg,
	}, nil
}

// AutoMigrate 自动迁移
func (db *DB) AutoMigrate(dst ...interface{}) error {
	return db.db.AutoMigrate(dst...)
}

// Create 创建记录
func (db *DB) Create(ctx context.Context, value interface{}) error {
	return db.db.WithContext(ctx).Create(value).Error
}

// First 查询第一条记录
func (db *DB) First(ctx context.Context, dst interface{}, conds ...interface{}) error {
	return db.db.WithContext(ctx).First(dst, conds...).Error
}

// Find 查询多条记录
func (db *DB) Find(ctx context.Context, dst interface{}, conds ...interface{}) error {
	return db.db.WithContext(ctx).Find(dst, conds...).Error
}

// Update 更新记录
func (db *DB) Update(ctx context.Context, model interface{}, column string, value interface{}) error {
	return db.db.WithContext(ctx).Model(model).Update(column, value).Error
}

// Updates 批量更新记录
func (db *DB) Updates(ctx context.Context, model interface{}, values interface{}) error {
	return db.db.WithContext(ctx).Model(model).Updates(values).Error
}

// Delete 删除记录
func (db *DB) Delete(ctx context.Context, model interface{}, conds ...interface{}) error {
	return db.db.WithContext(ctx).Delete(model, conds...).Error
}

// Where 条件查询
func (db *DB) Where(query interface{}, args ...interface{}) database.DB {
	return &DB{
		db:  db.db.Where(query, args...),
		cfg: db.cfg,
	}
}

// Begin 开启事务
func (db *DB) Begin(ctx context.Context, opts ...*database.TxOptions) (database.Tx, error) {
	var txOpts *gorm.Session
	if len(opts) > 0 {
		txOpts = &gorm.Session{}
		if opts[0].ReadOnly {
			txOpts.SkipDefaultTransaction = true
		}
	}

	tx := db.db.WithContext(ctx).Session(txOpts).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Tx{tx: tx}, nil
}

// Stats 获取连接池统计信息
func (db *DB) Stats() database.Stats {
	sqlDB, err := db.db.DB()
	if err != nil {
		return database.Stats{}
	}

	stats := sqlDB.Stats()
	return database.Stats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DB 获取原生GORM实例
func (db *DB) DB() *gorm.DB {
	return db.db
}

// Tx MySQL事务实现
type Tx struct {
	tx *gorm.DB
}

// Create 创建记录
func (tx *Tx) Create(ctx context.Context, value interface{}) error {
	return tx.tx.WithContext(ctx).Create(value).Error
}

// First 查询第一条记录
func (tx *Tx) First(ctx context.Context, dst interface{}, conds ...interface{}) error {
	return tx.tx.WithContext(ctx).First(dst, conds...).Error
}

// Find 查询多条记录
func (tx *Tx) Find(ctx context.Context, dst interface{}, conds ...interface{}) error {
	return tx.tx.WithContext(ctx).Find(dst, conds...).Error
}

// Update 更新记录
func (tx *Tx) Update(ctx context.Context, model interface{}, column string, value interface{}) error {
	return tx.tx.WithContext(ctx).Model(model).Update(column, value).Error
}

// Updates 批量更新记录
func (tx *Tx) Updates(ctx context.Context, model interface{}, values interface{}) error {
	return tx.tx.WithContext(ctx).Model(model).Updates(values).Error
}

// Delete 删除记录
func (tx *Tx) Delete(ctx context.Context, model interface{}, conds ...interface{}) error {
	return tx.tx.WithContext(ctx).Delete(model, conds...).Error
}

// Where 条件查询
func (tx *Tx) Where(query interface{}, args ...interface{}) database.Tx {
	return &Tx{tx: tx.tx.Where(query, args...)}
}

// Commit 提交事务
func (tx *Tx) Commit() error {
	return tx.tx.Commit().Error
}

// Rollback 回滚事务
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback().Error
}

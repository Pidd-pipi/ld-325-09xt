package repository

import "gorm.io/gorm"

// TxManager 让 service 层在不依赖具体 ORM 细节的前提下编排数据库事务。
// 回调内构造的仓储全部绑定到同一事务句柄，保证“报价改写 + 历史写入 +
// 预警触发 + 站内消息”原子提交。
type TxManager struct{ db *gorm.DB }

func NewTxManager(db *gorm.DB) *TxManager { return &TxManager{db: db} }

func (m *TxManager) RunInTx(fn func(tx *gorm.DB) error) error {
	return m.db.Transaction(fn)
}

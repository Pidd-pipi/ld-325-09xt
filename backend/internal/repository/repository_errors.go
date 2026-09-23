package repository

import "strings"

// isDuplicateKey 识别 PostgreSQL 与 SQLite 的唯一约束冲突，
// 用于把数据库级并发保护转译成业务冲突。
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

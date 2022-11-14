package action

// ActionHandler 执行动作钩子
type ActionHandler interface {
	// Pre 执行前准备
	Pre() error

	// Hook 执行
	Hook() error

	// Post 执行后清理 (无论执行是否成功，都应该有Post的清理)
	Post() error
}

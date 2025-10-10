package goal

import "time"

type Goal struct {
	Name       string
	IsAchieved func(ctx *Context) bool
	Ensure     func(ctx *Context) error
}
type GoalSet struct {
	Name       string
	Goals      []Goal
	MaxRetries int
	Timeout    time.Duration
}

var Registry = map[string]*GoalSet{}

// GoalSetBuilder 构建器
type GoalSetBuilder struct {
	name       string
	goals      []Goal
	maxRetries int
	timeout    time.Duration
}

func NewGoalSetBuilder(name string) *GoalSetBuilder {
	return &GoalSetBuilder{
		name:       name,
		goals:      make([]Goal, 0),
		maxRetries: 0, // 默认不重试
		timeout:    10 * time.Second,
	}
}

func (b *GoalSetBuilder) AddGoal(g Goal) *GoalSetBuilder {
	b.goals = append(b.goals, g)
	return b
}

func (b *GoalSetBuilder) WithMaxRetries(n int) *GoalSetBuilder {
	b.maxRetries = n
	return b
}

func (b *GoalSetBuilder) WithTimeout(d time.Duration) *GoalSetBuilder {
	b.timeout = d
	return b
}

func (b *GoalSetBuilder) BuildAndRegister() {

	Registry[b.name] = &GoalSet{
		Name:       b.name,
		Goals:      b.goals,
		MaxRetries: b.maxRetries,
		Timeout:    b.timeout,
	}
}

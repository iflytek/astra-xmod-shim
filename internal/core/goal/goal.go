// pkg/goal/goal.go

package goal

type Goal struct {
	Name       string
	IsAchieved func(ctx *Context) bool
	Ensure     func(ctx *Context) error
}

var Registry = map[string]*GoalSet{}

package reconciler

import (
	"modserv-shim/internal/core/goal"
	dto "modserv-shim/internal/dto/deploy"
)

type Reconciler struct {
}

func NewReconciler() *Reconciler {
	return &Reconciler{}
}

func (o *Reconciler) Reconcile(spec *dto.DeploySpec) error {

	// 组装 ctx
	goalSetCtx := &goal.Context{
		Shimlet:    spec.Shimlet,
		ResourceId: spec.ServiceId,
		Data:       make(map[string]any),
	}

	// run goals
	for _, singleGoal := range spec.GoalSet.Goals {
		if !singleGoal.IsAchieved(goalSetCtx) {
			err := singleGoal.Ensure(goalSetCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

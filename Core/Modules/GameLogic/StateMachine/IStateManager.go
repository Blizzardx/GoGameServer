package StateMachine

type IStateManager interface {
	OnInit()
	OnChangeStatus_Before(fromStatus IState, toStatus IState)
	OnChangeStatus_After(fromStatus IState, toStatus IState)
	OnTick_Before()
	OnTick_After()
}

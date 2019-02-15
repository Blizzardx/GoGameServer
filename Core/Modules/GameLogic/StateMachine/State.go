package StateMachine

type stateLogic struct {
	logic IState
}

func createState(state IState) *stateLogic {
	stateInfo := &stateLogic{}
	stateInfo.init(state)
	return stateInfo
}
func (stateLogic *stateLogic) init(state IState) {
	stateLogic.logic = state
	stateLogic.logic.OnInit()
}
func (stateLogic *stateLogic) enter() {
	stateLogic.logic.OnEnter()
}
func (stateLogic *stateLogic) exit() {
	stateLogic.logic.OnExit()
}
func (stateLogic *stateLogic) tick() {
	stateLogic.logic.OnTick()
}

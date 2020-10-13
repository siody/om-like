package timerwheel

//Node is a job put into wheel.
//Empty node next and previous should point to itself.
//Key and value is related to snapshot. Key will be key of map.
//VariableTime is a nano based timestamp, deciding which bucket it
//will locate to.
//Active function triggers your job, it will be execute in wheel main
//routine, so you may go anther routine to finish your job.
type Node interface {
	GetVariableTime() int64
	SetPreviousInVariableOrder(Node)
	SetNextInVariableOrder(Node)
	GetPreviousInVariableOrder() Node
	GetNextInVariableOrder() Node
	Active()
	GetKey() interface{}
	GetValue() interface{}
}

type sentinel struct {
	n, p Node
}

func newSentinel() Node {
	this := new(sentinel)
	this.n = this
	this.p = this
	return this
}

func (s *sentinel) SetPreviousInVariableOrder(node Node) {
	s.p = node
}

func (s *sentinel) SetNextInVariableOrder(node Node) {
	s.n = node
}

func (s *sentinel) GetPreviousInVariableOrder() Node {
	return s.p
}

func (s *sentinel) GetNextInVariableOrder() Node {
	return s.n
}

func (s *sentinel) GetVariableTime() int64 {
	return 0
}

func (s *sentinel) Active() {}

func (s *sentinel) GetKey() interface{} {
	return nil
}

func (s *sentinel) GetValue() interface{} {
	return nil
}

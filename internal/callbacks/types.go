package callbacks

const (
	ActionApprove = "approve"
	ActionDecline = "decline"
)

type Data interface {
	Action() string
}

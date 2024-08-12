package input

type (
	FocusEvent struct{}
	BlurEvent  struct{}
)

func (FocusEvent) String() string { return "focus" }
func (BlurEvent) String() string  { return "blur" }

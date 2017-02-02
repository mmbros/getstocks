package run

type set map[string]struct{}

func NewSet() set {
	return map[string]struct{}{}
}

func (s set) Add(item string) bool {
	if _, ok := s[item]; ok {
		// already exists
		return false
	}
	s[item] = struct{}{}
	return true
}

func (s set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

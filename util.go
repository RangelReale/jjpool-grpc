package jjpool_grpc

//
// UniqueStringList
//

type UniqueStringList []string

func (l UniqueStringList) Add(value ...string) {
	for _, addr := range value {
		exists := false
		for _, a := range l {
			if a == addr {
				exists = true
				break
			}
		}
		if !exists {
			l = append(l, addr)
		}
	}
}

func (l *UniqueStringList) Remove(value ...string) {
	var newlist []string

	for _, addr := range value {
		exists := false
		for _, a := range *l {
			if a == addr {
				exists = true
				break
			}
		}
		if !exists {
			newlist = append(newlist, addr)
		}
	}
	*l = newlist
}

package humus

import "unsafe"

type unsafeSlice []byte

//Honestly, is this a good idea?
//For traversing the tree we constantly build up the current path.
//For convenience, using strings as an input is in my opinion the best.
//However, strings are immutable. Therefore, to compute the current path we have
//to allocate a new string. This is a big bottleneck. Right now, use a byte slice to
//calculate a string for key-ing the position map. Strings.Builder cannot do this as we
//have to backtrack.
//TODO: Does go optimize map[string(value)] where value is a byte slice? If so, this is unnecessary.
func (u unsafeSlice) pred() Predicate {
	return *(*Predicate)(unsafe.Pointer(&u))
}

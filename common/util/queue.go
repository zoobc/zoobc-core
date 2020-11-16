package util

// SimpleQueueAddElement add an element at the rear of the queue (last element of slice) and pop the one in front (first element of slice)
func SimpleQueueAddElement(a []interface{}, b interface{}) (poppedElement interface{}, newSlice []interface{}) {
	newSlice = append(a[1:], b)
	poppedElement = a[0]
	return
}

package resource

func Sort(i, j Resource) int {
	switch {
	case i.GetID() < j.GetID():
		return -1
	case i.GetID() > j.GetID():
		return 1
	default:
		return 0
	}
}
